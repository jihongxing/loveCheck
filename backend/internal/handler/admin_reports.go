package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"lovecheck/internal/db"
	"lovecheck/internal/model"
)

func parseAdminPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func parseTagJSON(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var tags []string
	if err := json.Unmarshal([]byte(raw), &tags); err == nil {
		return tags
	}
	return []string{raw}
}

func parseEvidenceJSON(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var evidences []string
	if err := json.Unmarshal([]byte(raw), &evidences); err == nil {
		return evidences
	}
	return []string{raw}
}

func HandleAdminListReports(c *gin.Context) {
	kind := strings.TrimSpace(c.DefaultQuery("kind", "person"))
	status := strings.TrimSpace(c.DefaultQuery("status", "active"))
	search := strings.TrimSpace(c.Query("q"))
	page, pageSize := parseAdminPagination(c)
	offset := (page - 1) * pageSize

	switch kind {
	case "person":
		listAdminRiskReports(c, status, search, page, pageSize, offset)
	case "company":
		listAdminCompanyReports(c, status, search, page, pageSize, offset)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_kind"})
	}
}

func listAdminRiskReports(c *gin.Context, status, search string, page, pageSize, offset int) {
	query := db.DB.Model(&model.RiskRecord{})
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("display_name ILIKE ? OR location_city ILIKE ? OR description ILIKE ?", like, like, like)
	}

	var total int64
	query.Count(&total)

	var records []model.RiskRecord
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records)

	hashes := make([]string, 0, len(records))
	seen := make(map[string]bool)
	for _, rec := range records {
		if rec.TargetHash != "" && !seen[rec.TargetHash] {
			seen[rec.TargetHash] = true
			hashes = append(hashes, rec.TargetHash)
		}
	}

	statsMap := make(map[string]model.TargetStats)
	if len(hashes) > 0 {
		var stats []model.TargetStats
		db.DB.Where("target_hash IN ?", hashes).Find(&stats)
		for _, stat := range stats {
			statsMap[stat.TargetHash] = stat
		}
	}

	items := make([]gin.H, 0, len(records))
	for _, rec := range records {
		stat, hasAppeal := statsMap[rec.TargetHash]
		items = append(items, gin.H{
			"id":                    rec.ID,
			"kind":                  "person",
			"status":                rec.Status,
			"display_name":          rec.DisplayName,
			"location_city":         rec.LocationCity,
			"description":           rec.Description,
			"tags":                  parseTagJSON(rec.Tags),
			"evidences":             SignEvidenceURLs(parseEvidenceJSON(rec.EvidenceMaskURL)),
			"created_at":            rec.CreatedAt,
			"reporter_display_name": resolveReporterDisplayName(&rec),
			"reporter_city":         strings.TrimSpace(rec.ReporterCity),
			"verification_level":    inferVerificationLevel(&rec),
			"has_appeal":            hasAppeal,
			"appeal_reason":         stat.AppealReason,
			"reporter_votes":        stat.ReporterVotes,
			"appeal_votes":          stat.AppealVotes,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     items,
		"kind":      "person",
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

func listAdminCompanyReports(c *gin.Context, status, search string, page, pageSize, offset int) {
	query := db.DB.Model(&model.CompanyRecord{})
	if status != "" && status != "all" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("display_name ILIKE ? OR company_name ILIKE ? OR registration_no ILIKE ? OR description ILIKE ?", like, like, like, like)
	}

	var total int64
	query.Count(&total)

	var records []model.CompanyRecord
	query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records)

	hashes := make([]string, 0, len(records))
	seen := make(map[string]bool)
	for _, rec := range records {
		if rec.CompanyHash != "" && !seen[rec.CompanyHash] {
			seen[rec.CompanyHash] = true
			hashes = append(hashes, rec.CompanyHash)
		}
	}

	statsMap := make(map[string]model.CompanyStats)
	if len(hashes) > 0 {
		var stats []model.CompanyStats
		db.DB.Where("company_hash IN ?", hashes).Find(&stats)
		for _, stat := range stats {
			statsMap[stat.CompanyHash] = stat
		}
	}

	items := make([]gin.H, 0, len(records))
	for _, rec := range records {
		stat, hasAppeal := statsMap[rec.CompanyHash]
		items = append(items, gin.H{
			"id":                    rec.ID,
			"kind":                  "company",
			"status":                rec.Status,
			"display_name":          rec.DisplayName,
			"registration_no":       rec.RegistrationNo,
			"industry":              rec.Industry,
			"location_city":         rec.LocationCity,
			"description":           rec.Description,
			"tags":                  parseTagJSON(rec.Tags),
			"evidences":             SignEvidenceURLs(parseEvidenceJSON(rec.EvidenceMaskURL)),
			"created_at":            rec.CreatedAt,
			"reporter_display_name": rec.ReporterDisplayName,
			"reporter_city":         strings.TrimSpace(rec.ReporterCity),
			"verification_level":    rec.VerificationLevel,
			"employment_period":     rec.EmploymentPeriod,
			"position":              rec.Position,
			"has_appeal":            hasAppeal,
			"appeal_reason":         stat.AppealReason,
			"reporter_votes":        stat.ReporterVotes,
			"company_votes":         stat.CompanyVotes,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"items":     items,
		"kind":      "company",
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	})
}

func HandleAdminUpdateReportStatus(c *gin.Context) {
	kind := strings.TrimSpace(c.Param("kind"))
	id := strings.TrimSpace(c.Param("id"))
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
		return
	}

	status := strings.TrimSpace(req.Status)
	if status != "active" && status != "hidden" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status"})
		return
	}

	switch kind {
	case "person":
		var rec model.RiskRecord
		if db.DB.First(&rec, id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		db.DB.Model(&rec).Update("status", status)
		c.JSON(http.StatusOK, gin.H{"updated": true, "status": status})
	case "company":
		var rec model.CompanyRecord
		if db.DB.First(&rec, id).Error != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		db.DB.Model(&rec).Update("status", status)
		c.JSON(http.StatusOK, gin.H{"updated": true, "status": status})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_kind"})
	}
}
