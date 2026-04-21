package handler

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"lovecheck/internal/bloom"
	"lovecheck/internal/db"
	"lovecheck/internal/middleware"
	"lovecheck/internal/model"
	"lovecheck/internal/storage"
	"lovecheck/pkg/crypto"
	"lovecheck/pkg/filecheck"
	"lovecheck/pkg/logger"
	"lovecheck/pkg/reporteridentity"
	"lovecheck/pkg/scoring"
)

// maskCompanyName masks company name for privacy (e.g., "北京字节跳动科技有限公司" -> "北京字**科技有限公司")
func maskCompanyName(name string) string {
	runes := []rune(name)
	if len(runes) <= 4 {
		return name
	}
	// Mask middle characters, keep first 2 and last 2
	start := 2
	end := len(runes) - 2
	if end <= start {
		return name
	}
	for i := start; i < end && i < start+2; i++ {
		runes[i] = '*'
	}
	return string(runes)
}

// maskRegistrationNo masks company registration number (e.g., "91110108MA01234567" -> "9111****234567")
func maskRegistrationNo(regNo string) string {
	if len(regNo) <= 8 {
		return regNo
	}
	return regNo[:4] + "****" + regNo[len(regNo)-6:]
}

// generateCompanyHash creates a deterministic hash from registration number only.
// Registration number is the true unique identifier - company names can change.
func generateCompanyHash(registrationNo string) string {
	normalized := strings.TrimSpace(registrationNo)
	return crypto.DeterministicHash(normalized)
}

func companyAppealObjectPrefix(registrationNo string) string {
	registrationNo = strings.TrimSpace(registrationNo)
	if len(registrationNo) == 0 {
		return "unknown"
	}
	if len(registrationNo) <= 8 {
		return registrationNo
	}
	return registrationNo[:8]
}

func parseSubmittedTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	var tags []string
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &tags); err == nil {
			return compactTags(tags)
		}
	}

	return compactTags(strings.Split(raw, ","))
}

func compactTags(tags []string) []string {
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		out = append(out, tag)
	}
	return out
}

// HandleCompanyReport handles company abuse report submission
func HandleCompanyReport(c *gin.Context) {
	reporterPhone := c.PostForm("reporter_phone")
	companyName := c.PostForm("company_name")
	registrationNo := c.PostForm("registration_no")
	industry := c.PostForm("industry")
	locationCity := c.PostForm("location_city")
	tags := c.PostForm("tags")
	description := c.PostForm("description")
	employmentPeriod := c.PostForm("employment_period")
	position := c.PostForm("position")

	if reporterPhone == "" || registrationNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_required_fields"})
		return
	}

	// Company name is optional but recommended for display
	if companyName == "" {
		companyName = "未提供公司名称"
	}

	// Sanitize inputs
	if len([]rune(companyName)) > 200 {
		companyName = string([]rune(companyName)[:200])
	}
	if len([]rune(locationCity)) > 100 {
		locationCity = string([]rune(locationCity)[:100])
	}
	if len([]rune(description)) > 5000 {
		description = string([]rune(description)[:5000])
	}
	if len(tags) > 1000 {
		tags = tags[:1000]
	}
	if len([]rune(employmentPeriod)) > 100 {
		employmentPeriod = string([]rune(employmentPeriod)[:100])
	}
	if len([]rune(position)) > 100 {
		position = string([]rune(position)[:100])
	}

	// Convert comma-separated tags to JSON array
	tagsArray := parseSubmittedTags(tags)
	tagsJSON, _ := json.Marshal(tagsArray)

	// Handle evidence file uploads
	const maxFileSize int64 = 10 << 20
	allowedTypes := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true,
		"image/webp": true, "image/bmp": true, "application/pdf": true,
	}

	evidenceMaskURLs := []string{} // Initialize as empty array, not nil
	if form, err := c.MultipartForm(); err == nil {
		files := form.File["evidence_files[]"]
		if len(files) > 9 {
			files = files[:9]
		}
		for _, file := range files {
			if file.Size > maxFileSize {
				continue
			}
			ct := file.Header.Get("Content-Type")
			if !allowedTypes[ct] {
				continue
			}
			openedFile, err := file.Open()
			if err != nil {
				continue
			}
			// Validate file magic bytes
			if !filecheck.ValidateMagicBytes(openedFile, ct) {
				openedFile.Close()
				continue
			}
			if seeker, ok := openedFile.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
			} else {
				openedFile.Close()
				continue
			}
			objectName := buildUniqueEvidenceObjectName("comp", crypto.MaskName(companyName), file.Filename)

			uploadedName, err := storage.UploadEvidence(objectName, openedFile, file.Size, ct)
			openedFile.Close()
			if err != nil {
				logger.Log.Error().Err(err).Str("file", file.Filename).Msg("MinIO upload failed")
				continue
			}
			evidenceMaskURLs = append(evidenceMaskURLs, uploadedName)
		}
	}
	evidenceJSON, _ := json.Marshal(evidenceMaskURLs)

	reporterHash := crypto.DeterministicHash(reporterPhone)
	companyHash := generateCompanyHash(registrationNo)
	maskedCompanyName := maskCompanyName(companyName)
	maskedRegNo := maskRegistrationNo(registrationNo)

	vLevel := 1
	if len(evidenceMaskURLs) > 0 {
		vLevel = 3
	}

	record := model.CompanyRecord{
		CompanyHash:         companyHash,
		CompanyName:         companyName,
		DisplayName:         maskedCompanyName,
		RegistrationNo:      maskedRegNo,
		Industry:            industry,
		LocationCity:        locationCity,
		Description:         description,
		RiskLevel:           2,
		Tags:                string(tagsJSON),
		EvidenceMaskURL:     string(evidenceJSON),
		Status:              "active",
		ReporterHash:        reporterHash,
		ReporterDisplayName: reporteridentity.NicknameFromHash(reporterHash),
		VerificationLevel:   vLevel,
		ReporterCity:        trustedReporterRegion(c),
		EmploymentPeriod:    employmentPeriod,
		Position:            position,
	}

	res := db.DB.Create(&record)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error"})
		return
	}

	bloom.Add(companyHash)

	c.JSON(http.StatusOK, gin.H{
		"message":   "company_report_submitted",
		"record_id": record.ID,
	})
}

// HandleCompanyQuery performs company lookup by registration number
func HandleCompanyQuery(c *gin.Context) {
	registrationNo := c.Query("registration_no")

	if registrationNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "registration_no_required"})
		return
	}

	companyHash := generateCompanyHash(registrationNo)

	// Bloom filter fast-path
	if !bloom.MayExist(companyHash) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "clean",
			"message": "No reports found for this company.",
		})
		return
	}

	// Query database
	var matchedRecords []model.CompanyRecord
	db.DB.Where("company_hash = ? AND status = ?", companyHash, "active").
		Order("created_at DESC").
		Limit(maxQueryResults).
		Find(&matchedRecords)

	if len(matchedRecords) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "clean",
			"message": "No reports found for this company.",
		})
		return
	}

	// Aggregate data
	uniqueTags := make(map[string]bool)
	uniqueLocations := make(map[string]bool)
	var finalEvidences []string
	var descriptions []string
	displayName := matchedRecords[0].DisplayName

	for _, rec := range matchedRecords {
		if strings.TrimSpace(rec.LocationCity) != "" {
			uniqueLocations[rec.LocationCity] = true
		}
		if strings.TrimSpace(rec.Description) != "" {
			descriptions = append(descriptions, rec.Description)
		}
		var parsedTags []string
		if err := json.Unmarshal([]byte(rec.Tags), &parsedTags); err == nil {
			for _, tag := range parsedTags {
				uniqueTags[tag] = true
			}
		}
		if rec.EvidenceMaskURL != "" {
			var parsedEvs []string
			if err := json.Unmarshal([]byte(rec.EvidenceMaskURL), &parsedEvs); err == nil {
				finalEvidences = append(finalEvidences, parsedEvs...)
			}
		}
	}

	var locArray []string
	for loc := range uniqueLocations {
		locArray = append(locArray, loc)
	}
	var tagArray []string
	for tag := range uniqueTags {
		tagArray = append(tagArray, tag)
	}

	reportCount := len(matchedRecords)
	scoreBreakdown := scoring.Calculate(tagArray, reportCount)
	consensusRiskScore := scoreBreakdown.RiskScore

	queryToken := companyHash

	firstReportAt := matchedRecords[0].CreatedAt
	latestReportAt := matchedRecords[0].CreatedAt
	reportDates := make([]time.Time, 0, len(matchedRecords))
	for _, rec := range matchedRecords {
		reportDates = append(reportDates, rec.CreatedAt)
		if rec.CreatedAt.Before(firstReportAt) {
			firstReportAt = rec.CreatedAt
		}
		if rec.CreatedAt.After(latestReportAt) {
			latestReportAt = rec.CreatedAt
		}
	}

	// Check for company appeal
	var stats model.CompanyStats
	hasAppeal := false
	if err := db.DB.Where("company_hash = ?", queryToken).First(&stats).Error; err == nil {
		hasAppeal = true
	}

	if hasAppeal {
		den := float64(stats.ReporterVotes) + float64(stats.CompanyVotes)
		dampingRatio := 1.0
		if den > 0 {
			dampingRatio = float64(stats.ReporterVotes) / den
		}
		consensusRiskScore = math.Round(consensusRiskScore*dampingRatio*10) / 10
		scoreBreakdown.RiskScore = consensusRiskScore
	}

	var appealEvs []string
	if hasAppeal && stats.AppealEvidence != "" {
		_ = json.Unmarshal([]byte(stats.AppealEvidence), &appealEvs)
	}

	var appealAt *time.Time
	if hasAppeal && !stats.CreatedAt.IsZero() {
		t := stats.CreatedAt
		appealAt = &t
	}

	records := make([]gin.H, 0, len(matchedRecords))
	for _, rec := range matchedRecords {
		var recTags []string
		if err := json.Unmarshal([]byte(rec.Tags), &recTags); err != nil {
			recTags = []string{}
		}
		var recEvs []string
		if rec.EvidenceMaskURL != "" {
			if err := json.Unmarshal([]byte(rec.EvidenceMaskURL), &recEvs); err != nil {
				recEvs = []string{rec.EvidenceMaskURL}
			}
		}
		records = append(records, gin.H{
			"id":                    rec.ID,
			"display_name":          rec.DisplayName,
			"registration_no":       rec.RegistrationNo,
			"industry":              rec.Industry,
			"location_city":         rec.LocationCity,
			"description":           rec.Description,
			"tags":                  recTags,
			"evidences":             SignEvidenceURLs(recEvs),
			"employment_period":     rec.EmploymentPeriod,
			"position":              rec.Position,
			"created_at":            rec.CreatedAt,
			"reporter_display_name": rec.ReporterDisplayName,
			"verification_level":    rec.VerificationLevel,
			"reporter_city":         strings.TrimSpace(rec.ReporterCity),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "warning",
		"message":     "Caution: Multiple abuse reports found for this company!",
		"query_token": queryToken,
		"records":     records,
		"aggregated_profile": gin.H{
			"display_name":              displayName,
			"total_independent_reports": reportCount,
			"consensus_risk_score":      consensusRiskScore,
			"risk_level":                scoreBreakdown.RiskLevel,
			"score_breakdown":           scoreBreakdown,
			"active_cities":             locArray,
			"consolidated_tags":         tagArray,
			"descriptions":              descriptions,
			"evidences":                 SignEvidenceURLs(finalEvidences),
			"first_report_at":           firstReportAt,
			"latest_report_at":          latestReportAt,
			"report_dates":              reportDates,
			"has_appeal":                hasAppeal,
			"appeal_reason":             stats.AppealReason,
			"appeal_evidences":          SignEvidenceURLs(appealEvs),
			"appeal_at":                 appealAt,
			"reporter_votes":            stats.ReporterVotes,
			"company_votes":             stats.CompanyVotes,
		},
	})
}

// HandleCompanyAppeal handles company appeal submission
func HandleCompanyAppeal(c *gin.Context) {
	contactPhone := c.PostForm("contact_phone")
	registrationNo := c.PostForm("registration_no")
	reason := c.PostForm("reason")

	if contactPhone == "" || registrationNo == "" || reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_required_fields"})
		return
	}

	if len([]rune(reason)) > 5000 {
		reason = string([]rune(reason)[:5000])
	}

	companyHash := generateCompanyHash(registrationNo)

	// Handle evidence uploads
	appealAllowedTypes := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true,
		"image/webp": true, "image/bmp": true, "application/pdf": true,
	}
	evidenceMaskURLs := []string{} // Initialize as empty array, not nil
	if form, err := c.MultipartForm(); err == nil {
		files := form.File["evidence_files[]"]
		if len(files) > 9 {
			files = files[:9]
		}
		for _, file := range files {
			if file.Size > 10<<20 {
				continue
			}
			ct := file.Header.Get("Content-Type")
			if !appealAllowedTypes[ct] {
				continue
			}
			openedFile, err := file.Open()
			if err != nil {
				continue
			}
			if !filecheck.ValidateMagicBytes(openedFile, ct) {
				openedFile.Close()
				continue
			}
			if seeker, ok := openedFile.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
			} else {
				openedFile.Close()
				continue
			}
			objectName := buildUniqueEvidenceObjectName("comp_apl", companyAppealObjectPrefix(registrationNo), file.Filename)

			uploadedName, err := storage.UploadEvidence(objectName, openedFile, file.Size, ct)
			openedFile.Close()
			if err == nil {
				evidenceMaskURLs = append(evidenceMaskURLs, uploadedName)
			}
		}
	}

	evJson, _ := json.Marshal(evidenceMaskURLs)

	var stats model.CompanyStats
	res := db.DB.Where("company_hash = ?", companyHash).First(&stats)
	if res.Error != nil {
		stats = model.CompanyStats{
			CompanyHash:    companyHash,
			AppealReason:   reason,
			AppealEvidence: string(evJson),
			ReporterVotes:  0,
			CompanyVotes:   0,
		}
		db.DB.Create(&stats)
	} else {
		stats.AppealReason = reason
		stats.AppealEvidence = string(evJson)
		db.DB.Save(&stats)
	}

	c.JSON(http.StatusOK, gin.H{"message": "company_appeal_submitted"})
}

// HandleCompanyVote handles voting on company reports (reporter vs company appeal)
func HandleCompanyVote(c *gin.Context) {
	var req struct {
		RegistrationNo string `json:"registration_no"`
		Side           string `json:"side"`
		Fingerprint    string `json:"fingerprint"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_vote_payload"})
		return
	}

	if req.RegistrationNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "registration_no_required"})
		return
	}

	companyHash := generateCompanyHash(req.RegistrationNo)
	clientIP := c.ClientIP()

	if middleware.RedisClient != nil {
		ctx := c.Request.Context()

		// Per-IP daily vote cap
		dailyKey := "company_vote_daily:" + clientIP
		dailyCount, _ := middleware.RedisClient.Incr(ctx, dailyKey).Result()
		if dailyCount == 1 {
			middleware.RedisClient.Expire(ctx, dailyKey, 24*time.Hour)
		}
		if dailyCount > 20 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "vote_limit_exceeded"})
			return
		}

		// IP-based dedup
		voteKeyIP := "company_vote:" + companyHash + ":" + clientIP
		if middleware.RedisClient.Exists(ctx, voteKeyIP).Val() > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "already_voted"})
			return
		}

		// Fingerprint-based dedup
		if req.Fingerprint != "" {
			fpHash := crypto.DeterministicHash(req.Fingerprint)
			voteKeyFP := "company_vote_fp:" + companyHash + ":" + fpHash
			if middleware.RedisClient.Exists(ctx, voteKeyFP).Val() > 0 {
				c.JSON(http.StatusConflict, gin.H{"error": "already_voted"})
				return
			}
		}
	}

	var stats model.CompanyStats
	if err := db.DB.Where("company_hash = ?", companyHash).First(&stats).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "appeal_not_found"})
		return
	}

	if req.Side == "reporter" {
		stats.ReporterVotes++
	} else if req.Side == "company" {
		stats.CompanyVotes++
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_side_parameter"})
		return
	}
	db.DB.Save(&stats)

	// Auto-cleanse if company appeal gets 2x reporter votes with at least 10 total votes
	cleansed := false
	totalVotes := stats.ReporterVotes + stats.CompanyVotes
	if totalVotes >= 10 && float64(stats.CompanyVotes) > float64(stats.ReporterVotes)*2.0 {
		db.DB.Model(&model.CompanyRecord{}).
			Where("company_hash = ?", companyHash).
			Update("status", "cleansed_by_jury")
		cleansed = true
	}

	// Mark as voted
	if middleware.RedisClient != nil {
		ctx := c.Request.Context()
		voteKeyIP := "company_vote:" + companyHash + ":" + clientIP
		middleware.RedisClient.Set(ctx, voteKeyIP, "1", 30*24*time.Hour)
		if req.Fingerprint != "" {
			fpHash := crypto.DeterministicHash(req.Fingerprint)
			voteKeyFP := "company_vote_fp:" + companyHash + ":" + fpHash
			middleware.RedisClient.Set(ctx, voteKeyFP, "1", 30*24*time.Hour)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "vote_recorded",
		"cleansed": cleansed,
	})
}

// HandleCompanyStats returns public statistics about company reports
func HandleCompanyStats(c *gin.Context) {
	var totalReports int64
	db.DB.Model(&model.CompanyRecord{}).Where("status = ?", "active").Count(&totalReports)

	var totalCompanies int64
	db.DB.Model(&model.CompanyRecord{}).
		Where("status = ?", "active").
		Distinct("company_hash").
		Count(&totalCompanies)

	// Top industries
	type IndustryCount struct {
		Industry string
		Count    int64
	}
	var industries []IndustryCount
	db.DB.Model(&model.CompanyRecord{}).
		Select("industry, COUNT(*) as count").
		Where("status = ? AND industry != ''", "active").
		Group("industry").
		Order("count DESC").
		Limit(10).
		Scan(&industries)

	c.JSON(http.StatusOK, gin.H{
		"total_reports":   totalReports,
		"total_companies": totalCompanies,
		"top_industries":  industries,
	})
}
