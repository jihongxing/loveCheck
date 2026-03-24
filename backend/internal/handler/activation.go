package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"lovecheck/internal/db"
	"lovecheck/internal/model"
)

func generateCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	raw := strings.ToUpper(hex.EncodeToString(b))
	return fmt.Sprintf("LT-%s-%s-%s", raw[0:4], raw[4:8], raw[8:12])
}

// HandleGenerateCodes creates a batch of activation codes (200 / 500 / 1000).
func HandleGenerateCodes(c *gin.Context) {
	countStr := c.DefaultQuery("count", "200")
	var count int
	fmt.Sscanf(countStr, "%d", &count)
	if count < 1 || count > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "count must be 1-1000"})
		return
	}

	codes := make([]string, 0, count)
	for i := 0; i < count; i++ {
		code := generateCode()
		record := model.ActivationCode{
			Code:   code,
			Status: "unused",
		}
		if err := db.DB.Create(&record).Error; err != nil {
			continue
		}
		codes = append(codes, code)
	}

	c.JSON(http.StatusOK, gin.H{
		"generated": len(codes),
		"codes":     codes,
	})
}

// HandleActivate validates an activation code and binds it to a target hash.
func HandleActivate(c *gin.Context) {
	var req struct {
		Code       string `json:"code"`
		TargetHash string `json:"target_hash"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Code == "" || req.TargetHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_code_or_target"})
		return
	}

	code := strings.ToUpper(strings.TrimSpace(req.Code))

	var ac model.ActivationCode
	result := db.DB.Where("code = ?", code).First(&ac)
	if result.Error != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "invalid_code"})
		return
	}

	if ac.Status == "used" {
		if ac.TargetHash == req.TargetHash {
			if ac.ActivatedAt != nil && time.Since(*ac.ActivatedAt).Hours() < 24 {
				c.JSON(http.StatusOK, gin.H{"success": true, "message": "already_activated"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "code_already_used"})
		return
	}

	now := time.Now()
	db.DB.Model(&ac).Updates(map[string]interface{}{
		"status":       "used",
		"target_hash":  req.TargetHash,
		"activated_ip": c.ClientIP(),
		"activated_at": now,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "activated",
	})
}

// HandleCheckAccess checks if a target hash has been unlocked by a valid activation code.
func HandleCheckAccess(c *gin.Context) {
	targetHash := c.Query("target_hash")
	if targetHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"unlocked": false})
		return
	}

	var ac model.ActivationCode
	result := db.DB.Where("target_hash = ? AND status = ?", targetHash, "used").
		Order("activated_at DESC").First(&ac)
	if result.Error == nil {
		c.JSON(http.StatusOK, gin.H{"unlocked": true})
		return
	}

	var po model.PaymentOrder
	result = db.DB.Where("target_hash = ? AND status = ?", targetHash, "paid").First(&po)
	if result.Error == nil {
		c.JSON(http.StatusOK, gin.H{"unlocked": true})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unlocked": false})
}
