package handler

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	mrand "math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"lovecheck/internal/db"
	"lovecheck/internal/middleware"
	"lovecheck/internal/model"
)

func getAdminSecret() string {
	if s := os.Getenv("ADMIN_SECRET"); s != "" {
		return s
	}
	return "changeme_admin_secret"
}

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("X-Admin-Secret")
		if subtle.ConstantTimeCompare([]byte(secret), []byte(getAdminSecret())) != 1 {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// --- Purchase Platform CRUD ---

func HandleListPlatforms(c *gin.Context) {
	var platforms []model.PurchasePlatform
	db.DB.Order("sort_order ASC, id ASC").Find(&platforms)
	c.JSON(http.StatusOK, platforms)
}

func HandleCreatePlatform(c *gin.Context) {
	var p model.PurchasePlatform
	if err := c.ShouldBindJSON(&p); err != nil || p.Name == "" || p.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and url are required"})
		return
	}
	p.Active = true
	db.DB.Create(&p)
	c.JSON(http.StatusOK, p)
}

func HandleUpdatePlatform(c *gin.Context) {
	id := c.Param("id")
	var existing model.PurchasePlatform
	if db.DB.First(&existing, id).Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
		return
	}
	var update model.PurchasePlatform
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}
	db.DB.Model(&existing).Updates(map[string]interface{}{
		"name":       update.Name,
		"url":        update.URL,
		"icon":       update.Icon,
		"region":     update.Region,
		"sort_order": update.SortOrder,
		"active":     update.Active,
	})
	db.DB.First(&existing, id)
	c.JSON(http.StatusOK, existing)
}

func HandleDeletePlatform(c *gin.Context) {
	id := c.Param("id")
	db.DB.Delete(&model.PurchasePlatform{}, id)
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// Public: returns only active platforms (no auth required)
func HandlePublicPlatforms(c *gin.Context) {
	var platforms []model.PurchasePlatform
	db.DB.Where("active = ?", true).Order("sort_order ASC, id ASC").Find(&platforms)
	c.JSON(http.StatusOK, platforms)
}

// --- Public Stats (no auth, cached, with marketing offset) ---

func HandlePublicStats(c *gin.Context) {
	ctx := context.Background()
	cacheKey := "public_stats"

	if middleware.RedisClient != nil {
		if cached, err := middleware.RedisClient.Get(ctx, cacheKey).Result(); err == nil {
			c.Data(http.StatusOK, "application/json", []byte(cached))
			return
		}
	}

	var totalReports int64
	db.DB.Model(&model.RiskRecord{}).Count(&totalReports)

	var distinctCities int64
	db.DB.Model(&model.RiskRecord{}).
		Where("location_city != ''").
		Distinct("location_city").
		Count(&distinctCities)

	var usedCodes int64
	db.DB.Model(&model.ActivationCode{}).Where("status = ?", "used").Count(&usedCodes)

	const reportOffset = 20000
	jitter := int64(mrand.Intn(801) + 200) // 200~1000
	displayReports := totalReports + reportOffset + jitter
	displayReports = displayReports / 100 * 100

	displayAlerts := usedCodes + int64(float64(reportOffset)*0.25)
	displayAlerts = displayAlerts / 100 * 100

	if distinctCities < 10 {
		distinctCities = 10
	}

	result := gin.H{
		"reports": displayReports,
		"cities":  distinctCities,
		"alerts":  displayAlerts,
	}

	if middleware.RedisClient != nil {
		jsonBytes, _ := json.Marshal(result)
		middleware.RedisClient.Set(ctx, cacheKey, string(jsonBytes), 5*time.Minute)
	}

	c.JSON(http.StatusOK, result)
}

// HandleListUnusedCodes returns all unused activation codes for admin viewing.
func HandleListUnusedCodes(c *gin.Context) {
	var codes []model.ActivationCode
	db.DB.Where("status = ?", "unused").Order("created_at DESC").Find(&codes)

	type codeItem struct {
		Code      string    `json:"code"`
		CreatedAt time.Time `json:"created_at"`
	}
	items := make([]codeItem, 0, len(codes))
	for _, c := range codes {
		items = append(items, codeItem{Code: c.Code, CreatedAt: c.CreatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"codes": items, "total": len(items)})
}

// --- Dashboard Stats ---

func HandleDashboardStats(c *gin.Context) {
	ctx := context.Background()
	dashKey := "admin_dashboard_stats"
	if middleware.RedisClient != nil {
		if cached, err := middleware.RedisClient.Get(ctx, dashKey).Result(); err == nil {
			c.Data(http.StatusOK, "application/json", []byte(cached))
			return
		}
	}

	var totalReports, activeReports int64
	db.DB.Model(&model.RiskRecord{}).Count(&totalReports)
	db.DB.Model(&model.RiskRecord{}).Where("status = ?", "active").Count(&activeReports)

	var totalAppeals int64
	db.DB.Model(&model.TargetStats{}).Where("appeal_reason != ''").Count(&totalAppeals)

	var totalCodes, unusedCodes, usedCodes int64
	db.DB.Model(&model.ActivationCode{}).Count(&totalCodes)
	db.DB.Model(&model.ActivationCode{}).Where("status = ?", "unused").Count(&unusedCodes)
	db.DB.Model(&model.ActivationCode{}).Where("status = ?", "used").Count(&usedCodes)

	var totalPlatforms, activePlatforms int64
	db.DB.Model(&model.PurchasePlatform{}).Count(&totalPlatforms)
	db.DB.Model(&model.PurchasePlatform{}).Where("active = ?", true).Count(&activePlatforms)

	var recentActivations []model.ActivationCode
	db.DB.Where("status = ?", "used").Order("activated_at DESC").Limit(20).Find(&recentActivations)

	type activationItem struct {
		Code        string     `json:"code"`
		ActivatedAt *time.Time `json:"activated_at"`
		IP          string     `json:"ip"`
	}
	recent := make([]activationItem, 0, len(recentActivations))
	for _, a := range recentActivations {
		maskedCode := a.Code
		if len(maskedCode) > 7 {
			maskedCode = maskedCode[:7] + "****"
		}
		recent = append(recent, activationItem{
			Code:        maskedCode,
			ActivatedAt: a.ActivatedAt,
			IP:          a.ActivatedIP,
		})
	}

	// Revenue estimation
	price := 19.9
	estimatedRevenue := float64(usedCodes) * price

	// Reports in last 7 days
	var reportsLast7d int64
	db.DB.Model(&model.RiskRecord{}).Where("created_at > ?", time.Now().AddDate(0, 0, -7)).Count(&reportsLast7d)

	// Activations in last 7 days
	var activationsLast7d int64
	db.DB.Model(&model.ActivationCode{}).Where("status = ? AND activated_at > ?", "used", time.Now().AddDate(0, 0, -7)).Count(&activationsLast7d)

	result := gin.H{
		"reports": gin.H{
			"total":   totalReports,
			"active":  activeReports,
			"last_7d": reportsLast7d,
		},
		"appeals": gin.H{
			"total": totalAppeals,
		},
		"codes": gin.H{
			"total":   totalCodes,
			"unused":  unusedCodes,
			"used":    usedCodes,
			"last_7d": activationsLast7d,
		},
		"platforms": gin.H{
			"total":  totalPlatforms,
			"active": activePlatforms,
		},
		"revenue": gin.H{
			"estimated_total": estimatedRevenue,
			"currency":        "CNY",
		},
		"recent_activations": recent,
	}

	if middleware.RedisClient != nil {
		jsonBytes, _ := json.Marshal(result)
		middleware.RedisClient.Set(ctx, dashKey, string(jsonBytes), 30*time.Second)
	}

	c.JSON(http.StatusOK, result)
}
