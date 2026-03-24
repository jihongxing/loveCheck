package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"

	"lovecheck/internal/db"
	"lovecheck/internal/middleware"
	"lovecheck/internal/model"
	"lovecheck/pkg/crypto"
	"lovecheck/pkg/logger"
)

// HandleGetVAPIDKey returns the public VAPID key so the browser can subscribe.
func HandleGetVAPIDKey(c *gin.Context) {
	pub := os.Getenv("VAPID_PUBLIC_KEY")
	if pub == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "push_not_configured"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"public_key": pub})
}

// HandlePushSubscribe stores a browser push subscription for a watched phone.
// Limited to 5 subscriptions per IP per hour to prevent abuse.
func HandlePushSubscribe(c *gin.Context) {
	var req struct {
		Phone      string `json:"phone"`
		PhoneLocal string `json:"phone_local"`
		Endpoint   string `json:"endpoint"`
		KeyAuth    string `json:"key_auth"`
		KeyP256dh  string `json:"key_p256dh"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Phone == "" || req.Endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_subscription"})
		return
	}

	if len(req.Endpoint) > 2000 || len(req.KeyAuth) > 200 || len(req.KeyP256dh) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payload_too_large"})
		return
	}

	// Rate limit push subscriptions per IP
	if middleware.RedisClient != nil {
		ctx := c.Request.Context()
		subKey := "push_sub:" + c.ClientIP()
		count, _ := middleware.RedisClient.Incr(ctx, subKey).Result()
		if count == 1 {
			middleware.RedisClient.Expire(ctx, subKey, 1*time.Hour)
		}
		if count > 5 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "subscription_limit_exceeded"})
			return
		}
	}

	targetHash := crypto.DeterministicHash(req.Phone)

	var existing model.PushSubscription
	if db.DB.Where("endpoint = ? AND target_hash = ?", req.Endpoint, targetHash).First(&existing).Error == nil {
		c.JSON(http.StatusOK, gin.H{"message": "already_subscribed"})
		return
	}

	sub := model.PushSubscription{
		Endpoint:   req.Endpoint,
		KeyAuth:    req.KeyAuth,
		KeyP256dh:  req.KeyP256dh,
		TargetHash: targetHash,
	}
	db.DB.Create(&sub)
	c.JSON(http.StatusOK, gin.H{"message": "subscribed"})
}

// HandlePushUnsubscribe removes a push subscription.
func HandlePushUnsubscribe(c *gin.Context) {
	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	db.DB.Where("endpoint = ?", req.Endpoint).Delete(&model.PushSubscription{})
	c.JSON(http.StatusOK, gin.H{"message": "unsubscribed"})
}

// NotifyWatchers sends push notifications to all subscribers watching a given targetHash.
// Called internally after a new report is created.
func NotifyWatchers(targetHash string) {
	vapidPub := os.Getenv("VAPID_PUBLIC_KEY")
	vapidPriv := os.Getenv("VAPID_PRIVATE_KEY")
	vapidEmail := os.Getenv("VAPID_EMAIL")
	if vapidPub == "" || vapidPriv == "" {
		return
	}
	if vapidEmail == "" {
		vapidEmail = "admin@lovertrust.com"
	}

	var subs []model.PushSubscription
	db.DB.Where("target_hash = ?", targetHash).Find(&subs)
	if len(subs) == 0 {
		return
	}

	payload, _ := json.Marshal(map[string]string{
		"title": "LoverTrust Alert",
		"body":  "A new risk report has been filed for a number you are watching.",
		"url":   "/",
	})

	const maxWorkers = 10
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	for _, sub := range subs {
		wg.Add(1)
		sem <- struct{}{}
		go func(sub model.PushSubscription) {
			defer wg.Done()
			defer func() { <-sem }()

			s := &webpush.Subscription{
				Endpoint: sub.Endpoint,
				Keys: webpush.Keys{
					Auth:   sub.KeyAuth,
					P256dh: sub.KeyP256dh,
				},
			}
			resp, err := webpush.SendNotification(payload, s, &webpush.Options{
				Subscriber:      vapidEmail,
				VAPIDPublicKey:  vapidPub,
				VAPIDPrivateKey: vapidPriv,
				TTL:             3600,
			})
			if err != nil {
				epPreview := sub.Endpoint
				if len(epPreview) > 40 {
					epPreview = epPreview[:40]
				}
				logger.Log.Warn().Err(err).Str("endpoint", epPreview).Msg("Push delivery failed")
				if resp != nil && (resp.StatusCode == 404 || resp.StatusCode == 410) {
					db.DB.Delete(&sub)
				}
				return
			}
			resp.Body.Close()
		}(sub)
	}
	wg.Wait()
}
