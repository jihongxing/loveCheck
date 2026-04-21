package handler

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"lovecheck/internal/db"
	"lovecheck/internal/model"
	"lovecheck/pkg/crypto"
)

func normalizeAccessToken(token string) string {
	return strings.TrimSpace(token)
}

func hashAccessToken(token string) string {
	return crypto.DeterministicHash("access:" + normalizeAccessToken(token))
}

func generateAccessToken() string {
	buf := make([]byte, 24)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}

func issueAccessGrant(targetHash, source, sourceRef string, ttl time.Duration) (string, error) {
	token := generateAccessToken()
	grant := model.AccessGrant{
		TargetHash:      targetHash,
		ClientTokenHash: hashAccessToken(token),
		Source:          source,
		SourceRef:       sourceRef,
		Status:          "active",
	}
	if ttl > 0 {
		expiresAt := time.Now().Add(ttl)
		grant.ExpiresAt = &expiresAt
	}
	if err := db.DB.Create(&grant).Error; err != nil {
		return "", err
	}
	return token, nil
}

func grantFromToken(targetHash, token string) *model.AccessGrant {
	token = normalizeAccessToken(token)
	if token == "" || targetHash == "" {
		return nil
	}

	var grant model.AccessGrant
	now := time.Now()
	err := db.DB.
		Where("target_hash = ? AND client_token_hash = ? AND status = ?", targetHash, hashAccessToken(token), "active").
		Where("expires_at IS NULL OR expires_at > ?", now).
		Order("created_at DESC").
		First(&grant).Error
	if err != nil {
		return nil
	}
	return &grant
}

func requestAccessToken(c *gin.Context) string {
	if token := normalizeAccessToken(c.GetHeader("X-Access-Token")); token != "" {
		return token
	}
	if token := normalizeAccessToken(c.Query("access_token")); token != "" {
		return token
	}
	return ""
}

func hasUnlockedAccess(c *gin.Context, targetHash string) bool {
	return grantFromToken(targetHash, requestAccessToken(c)) != nil
}
