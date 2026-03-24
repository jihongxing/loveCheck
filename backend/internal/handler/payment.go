package handler

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"lovecheck/internal/db"
	"lovecheck/internal/model"
	"lovecheck/pkg/logger"
)

func getPayEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func generateOrderNo() string {
	ts := time.Now().Format("20060102150405")
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("LT%s%s", ts, strings.ToUpper(hex.EncodeToString(b)))
}

// xunhuSign computes the MD5 signature required by Xunhupay.
// 1. Sort params by key (ASCII), skip empty values and "hash"
// 2. Join as key1=value1&key2=value2...
// 3. Append APPSECRET (no &)
// 4. MD5 -> 32-char lowercase hex
func xunhuSign(params map[string]string, appSecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	first := true
	for _, k := range keys {
		if k == "hash" || params[k] == "" {
			continue
		}
		if !first {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(params[k])
		first = false
	}
	buf.WriteString(appSecret)

	sum := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(sum[:])
}

// HandlePayCreate creates a payment order for the requested provider (wechat/alipay/paypal).
func HandlePayCreate(c *gin.Context) {
	var req struct {
		TargetHash string `json:"target_hash"`
		Provider   string `json:"provider"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.TargetHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target_hash_required"})
		return
	}
	if req.Provider == "" {
		req.Provider = "wechat"
	}

	switch req.Provider {
	case "wechat", "alipay":
		handleXunhuPay(c, req.TargetHash, req.Provider)
	case "paypal":
		handlePayPalCreate(c, req.TargetHash)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_provider"})
	}
}

func handleXunhuPay(c *gin.Context, targetHash, provider string) {
	var appID, appSecret string
	if provider == "alipay" {
		appID = os.Getenv("XUNHU_ALI_APPID")
		appSecret = os.Getenv("XUNHU_ALI_APPSECRET")
	} else {
		appID = os.Getenv("XUNHU_APPID")
		appSecret = os.Getenv("XUNHU_APPSECRET")
	}
	if appID == "" || appSecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment_not_configured"})
		return
	}

	gateway := getPayEnv("XUNHU_GATEWAY", "https://api.xunhupay.com")
	notifyURL := os.Getenv("PAY_NOTIFY_URL")
	returnURL := getPayEnv("PAY_RETURN_URL", "/")
	price := getPayEnv("PAY_PRICE", "19.9")

	if notifyURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "notify_url_not_configured"})
		return
	}

	orderNo := generateOrderNo()

	order := model.PaymentOrder{
		OrderNo:    orderNo,
		TargetHash: targetHash,
		Provider:   provider,
		Amount:     price,
		Currency:   "CNY",
		Status:     "pending",
		ClientIP:   c.ClientIP(),
	}
	if err := db.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order_create_failed"})
		return
	}

	nonceStr := hex.EncodeToString(func() []byte { b := make([]byte, 16); rand.Read(b); return b }())

	payType := "WAP"
	if provider == "alipay" {
		payType = "jump"
	}

	params := map[string]string{
		"version":        "1.1",
		"appid":          appID,
		"trade_order_id": orderNo,
		"total_fee":      price,
		"title":          "LoverTrust Query Unlock",
		"time":           fmt.Sprintf("%d", time.Now().Unix()),
		"notify_url":     notifyURL,
		"return_url":     returnURL,
		"nonce_str":      nonceStr,
		"type":           payType,
		"wap_url":        returnURL,
		"wap_name":       "LoverTrust",
	}
	params["hash"] = xunhuSign(params, appSecret)

	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	resp, err := http.PostForm(gateway+"/payment/do.html", formData)
	if err != nil {
		logger.Log.Error().Err(err).Str("provider", provider).Msg("Xunhupay gateway request failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "gateway_unreachable"})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		ErrCode   int    `json:"errcode"`
		ErrMsg    string `json:"errmsg"`
		URL       string `json:"url"`
		URLQRCode string `json:"url_qrcode"`
		OpenID    string `json:"openid"`
		Hash      string `json:"hash"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Log.Error().Err(err).Str("body", string(body)).Msg("Xunhupay response parse failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "gateway_response_invalid"})
		return
	}

	if result.ErrCode != 0 {
		logger.Log.Warn().Int("errcode", result.ErrCode).Str("errmsg", result.ErrMsg).Msg("Xunhupay returned error")
		c.JSON(http.StatusBadGateway, gin.H{"error": "gateway_error", "detail": result.ErrMsg})
		return
	}

	if result.OpenID != "" {
		db.DB.Model(&order).Update("xunhu_order_id", result.OpenID)
	}

	c.JSON(http.StatusOK, gin.H{
		"order_no": orderNo,
		"pay_url":  result.URL,
		"qr_url":   result.URLQRCode,
		"amount":   price,
		"currency": "CNY",
		"provider": provider,
	})
}

// HandlePayNotify receives the payment callback from Xunhupay (covers both WeChat and Alipay).
// Xunhupay sends a form-encoded POST; we must respond with plain text "success".
func HandlePayNotify(c *gin.Context) {
	appSecret := os.Getenv("XUNHU_APPSECRET")
	aliSecret := os.Getenv("XUNHU_ALI_APPSECRET")
	if appSecret == "" && aliSecret == "" {
		c.String(http.StatusOK, "fail")
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	params := make(map[string]string)
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	receivedHash := params["hash"]
	if receivedHash == "" {
		c.String(http.StatusOK, "fail")
		return
	}

	verified := false
	if appSecret != "" && strings.EqualFold(receivedHash, xunhuSign(params, appSecret)) {
		verified = true
	}
	if !verified && aliSecret != "" && strings.EqualFold(receivedHash, xunhuSign(params, aliSecret)) {
		verified = true
	}
	if !verified {
		logger.Log.Warn().Str("received", receivedHash).Msg("Xunhupay callback signature mismatch")
		c.String(http.StatusOK, "fail")
		return
	}

	orderNo := params["trade_order_id"]
	status := params["status"]
	transactionID := params["transaction_id"]
	xunhuOrderID := params["open_order_id"]

	if orderNo == "" || status != "OD" {
		c.String(http.StatusOK, "success")
		return
	}

	var order model.PaymentOrder
	if err := db.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		logger.Log.Warn().Str("order_no", orderNo).Msg("Payment callback for unknown order")
		c.String(http.StatusOK, "success")
		return
	}

	if order.Status == "paid" {
		c.String(http.StatusOK, "success")
		return
	}

	now := time.Now()
	db.DB.Model(&order).Updates(map[string]interface{}{
		"status":         "paid",
		"transaction_id": transactionID,
		"xunhu_order_id": xunhuOrderID,
		"paid_at":        now,
	})

	logger.Log.Info().Str("order_no", orderNo).Str("provider", order.Provider).Str("amount", order.Amount).Msg("Payment confirmed")
	c.String(http.StatusOK, "success")
}

// HandlePayStatus lets the frontend poll for payment completion.
func HandlePayStatus(c *gin.Context) {
	orderNo := c.Query("order_no")
	if orderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_no_required"})
		return
	}

	var order model.PaymentOrder
	if err := db.DB.Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order_not_found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_no": order.OrderNo,
		"status":   order.Status,
		"paid":     order.Status == "paid",
	})
}

// --- PayPal integration ---

var (
	paypalTokenCache   string
	paypalTokenExpires time.Time
	paypalTokenMu      sync.Mutex
)

func paypalBaseURL() string {
	if getPayEnv("PAYPAL_MODE", "sandbox") == "live" {
		return "https://api-m.paypal.com"
	}
	return "https://api-m.sandbox.paypal.com"
}

func getPayPalAccessToken() (string, error) {
	paypalTokenMu.Lock()
	defer paypalTokenMu.Unlock()

	if paypalTokenCache != "" && time.Now().Before(paypalTokenExpires) {
		return paypalTokenCache, nil
	}

	clientID := os.Getenv("PAYPAL_CLIENT_ID")
	clientSecret := os.Getenv("PAYPAL_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return "", fmt.Errorf("paypal credentials not configured")
	}

	data := url.Values{"grant_type": {"client_credentials"}}
	req, _ := http.NewRequest("POST", paypalBaseURL()+"/v1/oauth2/token", strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", err
	}
	if tok.AccessToken == "" {
		return "", fmt.Errorf("paypal returned empty access_token")
	}

	paypalTokenCache = tok.AccessToken
	paypalTokenExpires = time.Now().Add(time.Duration(tok.ExpiresIn-60) * time.Second)
	return paypalTokenCache, nil
}

func handlePayPalCreate(c *gin.Context, targetHash string) {
	token, err := getPayPalAccessToken()
	if err != nil {
		logger.Log.Error().Err(err).Msg("PayPal access token failed")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment_not_configured"})
		return
	}

	price := getPayEnv("PAYPAL_PRICE", "2.99")
	currency := getPayEnv("PAYPAL_CURRENCY", "USD")
	returnURL := getPayEnv("PAY_RETURN_URL", "/")
	cancelURL := returnURL

	orderNo := generateOrderNo()

	order := model.PaymentOrder{
		OrderNo:    orderNo,
		TargetHash: targetHash,
		Provider:   "paypal",
		Amount:     price,
		Currency:   currency,
		Status:     "pending",
		ClientIP:   c.ClientIP(),
	}
	if err := db.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order_create_failed"})
		return
	}

	if strings.Contains(returnURL, "#") {
		returnURL = strings.Split(returnURL, "#")[0]
	}
	separator := "?"
	if strings.Contains(returnURL, "?") {
		separator = "&"
	}
	ppReturnURL := returnURL + separator + "paypal_capture=1&order_no=" + orderNo

	paypalOrder := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"reference_id": orderNo,
				"amount": map[string]string{
					"currency_code": currency,
					"value":         price,
				},
				"description": "LoverTrust Query Unlock",
			},
		},
		"application_context": map[string]string{
			"return_url": ppReturnURL,
			"cancel_url": cancelURL,
			"brand_name": "LoverTrust",
		},
	}

	body, _ := json.Marshal(paypalOrder)
	req, _ := http.NewRequest("POST", paypalBaseURL()+"/v2/checkout/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Error().Err(err).Msg("PayPal create order request failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "gateway_unreachable"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var ppResult struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Href   string `json:"href"`
			Rel    string `json:"rel"`
			Method string `json:"method"`
		} `json:"links"`
	}
	if err := json.Unmarshal(respBody, &ppResult); err != nil {
		logger.Log.Error().Err(err).Str("body", string(respBody)).Msg("PayPal response parse failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "gateway_response_invalid"})
		return
	}

	if ppResult.ID == "" {
		logger.Log.Warn().Str("body", string(respBody)).Msg("PayPal returned no order ID")
		c.JSON(http.StatusBadGateway, gin.H{"error": "gateway_error"})
		return
	}

	db.DB.Model(&order).Update("paypal_order_id", ppResult.ID)

	approveURL := ""
	for _, link := range ppResult.Links {
		if link.Rel == "approve" {
			approveURL = link.Href
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"order_no":    orderNo,
		"approve_url": approveURL,
		"amount":      price,
		"currency":    currency,
		"provider":    "paypal",
	})
}

// HandlePayPalCapture is called after the user approves on PayPal and is redirected back.
func HandlePayPalCapture(c *gin.Context) {
	var req struct {
		OrderNo string `json:"order_no"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OrderNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_no_required"})
		return
	}

	var order model.PaymentOrder
	if err := db.DB.Where("order_no = ? AND provider = ?", req.OrderNo, "paypal").First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order_not_found"})
		return
	}

	if order.Status == "paid" {
		c.JSON(http.StatusOK, gin.H{"paid": true, "order_no": order.OrderNo})
		return
	}

	if order.PayPalOrderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paypal_order_missing"})
		return
	}

	token, err := getPayPalAccessToken()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment_not_configured"})
		return
	}

	captureReq, _ := http.NewRequest("POST", paypalBaseURL()+"/v2/checkout/orders/"+order.PayPalOrderID+"/capture", bytes.NewReader([]byte("{}")))
	captureReq.Header.Set("Content-Type", "application/json")
	captureReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(captureReq)
	if err != nil {
		logger.Log.Error().Err(err).Msg("PayPal capture request failed")
		c.JSON(http.StatusBadGateway, gin.H{"error": "gateway_unreachable"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var captureResult struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &captureResult); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "gateway_response_invalid"})
		return
	}

	if captureResult.Status == "COMPLETED" {
		now := time.Now()
		db.DB.Model(&order).Updates(map[string]interface{}{
			"status":         "paid",
			"transaction_id": captureResult.ID,
			"paid_at":        now,
		})
		logger.Log.Info().Str("order_no", order.OrderNo).Str("paypal_id", captureResult.ID).Msg("PayPal payment captured")
		c.JSON(http.StatusOK, gin.H{"paid": true, "order_no": order.OrderNo})
		return
	}

	logger.Log.Warn().Str("status", captureResult.Status).Str("body", string(respBody)).Msg("PayPal capture not completed")
	c.JSON(http.StatusOK, gin.H{"paid": false, "order_no": order.OrderNo, "status": captureResult.Status})
}

