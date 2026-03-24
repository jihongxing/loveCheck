package handler

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"

	"lovecheck/internal/model"
	"lovecheck/pkg/reporteridentity"
)

// trustedReporterRegion returns a city/region label from trusted reverse-proxy
// headers only. Do not trust client-supplied form fields (spoofing).
// Set X-App-Geo-City from nginx geoip or similar; Cloudflare may send CF-IPCity.
func trustedReporterRegion(c *gin.Context) string {
	for _, key := range []string{"X-App-Geo-City", "CF-IPCity"} {
		v := strings.TrimSpace(c.GetHeader(key))
		if v == "" {
			continue
		}
		r := []rune(v)
		if len(r) > 80 {
			v = string(r[:80])
		}
		return v
	}
	return ""
}

func resolveReporterDisplayName(rec *model.RiskRecord) string {
	if s := strings.TrimSpace(rec.ReporterDisplayName); s != "" {
		return s
	}
	return reporteridentity.NicknameFromHash(rec.ReporterHash)
}

func inferVerificationLevel(rec *model.RiskRecord) int {
	if rec.VerificationLevel >= 1 && rec.VerificationLevel <= 3 {
		return rec.VerificationLevel
	}
	raw := strings.TrimSpace(rec.EvidenceMaskURL)
	if raw == "" || raw == "[]" {
		return 1
	}
	var ev []string
	if err := json.Unmarshal([]byte(raw), &ev); err == nil {
		if len(ev) > 0 {
			return 3
		}
		return 1
	}
	return 3
}
