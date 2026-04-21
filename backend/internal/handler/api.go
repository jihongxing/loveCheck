package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"

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

// collectHashes computes deterministic HMAC-SHA256 hashes for the given phone
// strings and returns a de-duplicated slice suitable for indexed DB lookups.
func collectHashes(phones ...string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, p := range phones {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		h := crypto.DeterministicHash(p)
		if !seen[h] {
			seen[h] = true
			out = append(out, h)
		}
	}
	return out
}

// queryByHashes performs two separate indexed queries (one per column) and
// merges results in Go. This guarantees PostgreSQL uses the Hash/Partial
// index on each column independently, avoiding OR-induced plan regressions.
const maxQueryResults = 500

func queryByHashes(hashes []string) []model.RiskRecord {
	idSet := make(map[uint]bool)
	var results []model.RiskRecord

	var batch []model.RiskRecord
	db.DB.Where("target_hash IN ? AND status = ?", hashes, "active").
		Order("created_at DESC").Limit(maxQueryResults).Find(&batch)
	for _, r := range batch {
		if !idSet[r.ID] {
			idSet[r.ID] = true
			results = append(results, r)
		}
	}

	if len(results) < maxQueryResults {
		batch = nil
		db.DB.Where("target_local_hash IN ? AND status = ?", hashes, "active").
			Order("created_at DESC").Limit(maxQueryResults - len(results)).Find(&batch)
		for _, r := range batch {
			if !idSet[r.ID] {
				idSet[r.ID] = true
				results = append(results, r)
			}
		}
	}

	return results
}

// findCanonicalHash looks up the first RiskRecord matching any of the provided
// hashes and returns its TargetHash (the canonical key for TargetStats).
func findCanonicalHash(hashes []string, fallbackPhone string) string {
	if len(hashes) > 0 {
		var rec model.RiskRecord
		if err := db.DB.Where("target_hash IN ?", hashes).First(&rec).Error; err == nil {
			return rec.TargetHash
		}
		if err := db.DB.Where("target_local_hash IN ?", hashes).First(&rec).Error; err == nil {
			return rec.TargetHash
		}
	}
	return crypto.DeterministicHash(fallbackPhone)
}

func lockedAggregatedProfile(displayName string, reportCount int, riskScore float64, riskLevel string, firstReportAt, latestReportAt time.Time) gin.H {
	return gin.H{
		"display_name":              displayName,
		"total_independent_reports": reportCount,
		"consensus_risk_score":      riskScore,
		"risk_level":                riskLevel,
		"first_report_at":           firstReportAt,
		"latest_report_at":          latestReportAt,
	}
}

// HandleReport stores both a full-phone hash (with dial code) and a local-phone
// hash (without dial code) for backward-compatible dual matching.
func HandleReport(c *gin.Context) {
	reporterPhone := c.PostForm("reporter_phone")
	targetPhone := c.PostForm("target_phone")
	targetPhoneLocal := c.PostForm("target_phone_local")
	targetName := c.PostForm("target_name")
	locationCity := c.PostForm("location_city")
	tags := c.PostForm("tags")
	description := c.PostForm("description")

	if reporterPhone == "" || targetPhone == "" || targetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_required_fields"})
		return
	}

	if strings.TrimSpace(reporterPhone) == strings.TrimSpace(targetPhone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "self_report_not_allowed"})
		return
	}

	if len([]rune(targetName)) > 50 {
		targetName = string([]rune(targetName)[:50])
	}
	if len([]rune(locationCity)) > 100 {
		locationCity = string([]rune(locationCity)[:100])
	}
	if len([]rune(description)) > 2000 {
		description = string([]rune(description)[:2000])
	}
	if len(tags) > 1000 {
		tags = tags[:1000]
	}

	const maxFileSize int64 = 10 << 20
	allowedTypes := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true,
		"image/webp": true, "image/bmp": true,
	}

	var evidenceMaskURLs []string
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
			// Validate file magic bytes to prevent Content-Type spoofing
			if !filecheck.ValidateMagicBytes(openedFile, ct) {
				openedFile.Close()
				continue
			}
			// Seek back to start after magic bytes check
			if seeker, ok := openedFile.(io.Seeker); ok {
				seeker.Seek(0, io.SeekStart)
			} else {
				openedFile.Close()
				continue
			}
			objectName := buildUniqueEvidenceObjectName("evd", crypto.MaskName(targetName), file.Filename)

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
	targetHash := crypto.DeterministicHash(targetPhone)
	maskedTargetName := crypto.MaskName(targetName)

	var targetLocalHash string
	if targetPhoneLocal != "" && targetPhoneLocal != targetPhone {
		targetLocalHash = crypto.DeterministicHash(targetPhoneLocal)
	}

	vLevel := 1
	if len(evidenceMaskURLs) > 0 {
		vLevel = 3
	}

	record := model.RiskRecord{
		TargetHash:          targetHash,
		TargetLocalHash:     targetLocalHash,
		DisplayName:         maskedTargetName,
		LocationCity:        locationCity,
		Description:         description,
		RiskLevel:           2,
		Tags:                tags,
		EvidenceMaskURL:     string(evidenceJSON),
		Status:              "active",
		ReporterHash:        reporterHash,
		ReporterDisplayName: reporteridentity.NicknameFromHash(reporterHash),
		VerificationLevel:   vLevel,
		ReporterCity:        trustedReporterRegion(c),
	}

	res := db.DB.Create(&record)
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error"})
		return
	}

	bloom.Add(targetHash)
	if targetLocalHash != "" {
		bloom.Add(targetLocalHash)
	}

	go NotifyWatchers(targetHash)

	c.JSON(http.StatusOK, gin.H{
		"message":   "report_submitted",
		"record_id": record.ID,
	})
}

// HandleQuery performs an O(1) indexed lookup using HMAC-SHA256 hashes.
// A Bloom Filter provides a fast-path rejection for hashes that definitely
// do not exist, eliminating unnecessary database I/O.
func HandleQuery(c *gin.Context) {
	searchPhone := c.Query("phone")
	searchLocal := c.Query("phone_local")

	hashes := collectHashes(searchPhone, searchLocal)
	if len(hashes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone_required"})
		return
	}

	// Bloom filter fast-path: if no hash could possibly exist, skip the DB entirely.
	if !bloom.MayExistAny(hashes) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "clean",
			"message": "0 hits found on engine. Pls stay reasonably vigilant regardless!",
		})
		return
	}

	// Two separate indexed queries, merged in Go — guarantees index usage.
	matchedRecords := queryByHashes(hashes)

	if len(matchedRecords) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  "clean",
			"message": "0 hits found on engine. Pls stay reasonably vigilant regardless!",
		})
		return
	}

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
		} else {
			uniqueTags[rec.Tags] = true
		}
		if rec.EvidenceMaskURL != "" {
			var parsedEvs []string
			if err := json.Unmarshal([]byte(rec.EvidenceMaskURL), &parsedEvs); err == nil {
				finalEvidences = append(finalEvidences, parsedEvs...)
			} else {
				finalEvidences = append(finalEvidences, rec.EvidenceMaskURL)
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

	queryToken := matchedRecords[0].TargetHash

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

	var stats model.TargetStats
	hasAppeal := false
	if err := db.DB.Where("target_hash = ?", queryToken).First(&stats).Error; err == nil {
		hasAppeal = true
	}

	if hasAppeal {
		// Avoid 0/0 -> NaN when no jury votes yet (JSON cannot encode NaN).
		den := float64(stats.ReporterVotes) + float64(stats.AppealVotes)
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

	unlocked := hasUnlockedAccess(c, queryToken)

	if !unlocked {
		c.JSON(http.StatusOK, gin.H{
			"status":          "warning",
			"message":         "Caution: We found multiple risk reports!",
			"query_token":     queryToken,
			"unlocked":        false,
			"access_required": true,
			"records":         []gin.H{},
			"aggregated_profile": lockedAggregatedProfile(
				displayName,
				reportCount,
				consensusRiskScore,
				scoreBreakdown.RiskLevel,
				firstReportAt,
				latestReportAt,
			),
		})
		return
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
			"location_city":         rec.LocationCity,
			"description":           rec.Description,
			"tags":                  recTags,
			"evidences":             SignEvidenceURLs(recEvs),
			"created_at":            rec.CreatedAt,
			"reporter_display_name": resolveReporterDisplayName(&rec),
			"verification_level":    inferVerificationLevel(&rec),
			"reporter_city":         strings.TrimSpace(rec.ReporterCity),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":          "warning",
		"message":         "Caution: We found multiple risk reports!",
		"query_token":     queryToken,
		"unlocked":        true,
		"access_required": false,
		"records":         records,
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
			"appeal_votes":              stats.AppealVotes,
		},
	})
}

// GetEvidence fetches object bytes from MinIO secure vault.
// Access requires a valid HMAC token to prevent unauthorized enumeration.
func GetEvidence(c *gin.Context) {
	filename := filepath.Base(c.Param("filename"))
	if filename == "" || filename == "." || filename == ".." {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing evidence identifier"})
		return
	}

	// Verify HMAC access token (generated by query results for unlocked users).
	token := c.Query("token")
	if !verifyEvidenceToken(filename, token) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access_denied"})
		return
	}

	if !strings.HasPrefix(filename, "evd_") && !strings.HasPrefix(filename, "apl_") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_filename"})
		return
	}

	ctx := context.Background()
	obj, err := storage.MinioClient.GetObject(ctx, storage.BucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evidence not found in cloud storage"})
		return
	}
	defer obj.Close()

	stat, err := obj.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot read object metadata"})
		return
	}

	c.DataFromReader(http.StatusOK, stat.Size, stat.ContentType, obj, map[string]string{
		"Cache-Control":       "private, max-age=3600",
		"Content-Disposition": "inline",
	})
}

// verifyEvidenceToken checks the HMAC-SHA256 token for evidence access.
// Token format: hex(HMAC-SHA256(filename + "|" + expiry, pepper))[0:16] + "." + expiry
func verifyEvidenceToken(filename, token string) bool {
	if token == "" {
		return false
	}
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	sig, expiryStr := parts[0], parts[1]
	var expiry int64
	if _, err := fmt.Sscanf(expiryStr, "%d", &expiry); err != nil {
		return false
	}
	if time.Now().Unix() > expiry {
		return false
	}
	expected := generateEvidenceToken(filename, expiry)
	expectedParts := strings.SplitN(expected, ".", 2)
	if len(expectedParts) != 2 {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(sig), []byte(expectedParts[0])) == 1
}

// generateEvidenceToken creates a time-limited HMAC token for evidence file access.
func generateEvidenceToken(filename string, expiry int64) string {
	data := fmt.Sprintf("%s|%d", filename, expiry)
	mac := hmac.New(sha256.New, []byte(crypto.GetEvidencePepper()))
	mac.Write([]byte(data))
	sig := hex.EncodeToString(mac.Sum(nil))[:16]
	return fmt.Sprintf("%s.%d", sig, expiry)
}

// SignEvidenceURLs adds HMAC tokens to evidence filenames for secure access.
func SignEvidenceURLs(filenames []string) []string {
	expiry := time.Now().Add(24 * time.Hour).Unix()
	signed := make([]string, 0, len(filenames))
	for _, f := range filenames {
		token := generateEvidenceToken(f, expiry)
		signed = append(signed, f+"?token="+token)
	}
	return signed
}

// HandleAppeal uses deterministic HMAC hashing for O(1) record lookup.
func HandleAppeal(c *gin.Context) {
	contactPhone := c.PostForm("contact_phone")
	targetPhone := c.PostForm("target_phone")
	targetPhoneLocal := c.PostForm("target_phone_local")
	reason := c.PostForm("reason")

	if contactPhone == "" || targetPhone == "" || reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_required_fields"})
		return
	}

	if len([]rune(reason)) > 5000 {
		reason = string([]rune(reason)[:5000])
	}

	hashes := collectHashes(targetPhone, targetPhoneLocal)
	canonicalHash := findCanonicalHash(hashes, targetPhone)

	appealAllowedTypes := map[string]bool{
		"image/jpeg": true, "image/png": true, "image/gif": true,
		"image/webp": true, "image/bmp": true,
	}
	var evidenceMaskURLs []string
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
			// Validate file magic bytes to prevent Content-Type spoofing
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
			namePrefix := targetPhone
			if len([]rune(namePrefix)) > 4 {
				namePrefix = string([]rune(namePrefix)[:4])
			}
			objectName := buildUniqueEvidenceObjectName("apl", crypto.MaskName(namePrefix), file.Filename)

			uploadedName, err := storage.UploadEvidence(objectName, openedFile, file.Size, ct)
			openedFile.Close()
			if err == nil {
				evidenceMaskURLs = append(evidenceMaskURLs, uploadedName)
			}
		}
	}

	evJson, _ := json.Marshal(evidenceMaskURLs)

	var stats model.TargetStats
	res := db.DB.Where("target_hash = ?", canonicalHash).First(&stats)
	if res.Error != nil {
		stats = model.TargetStats{
			TargetHash:     canonicalHash,
			AppealReason:   reason,
			AppealEvidence: string(evJson),
			ReporterVotes:  0,
			AppealVotes:    0,
		}
		db.DB.Create(&stats)
	} else {
		stats.AppealReason = reason
		stats.AppealEvidence = string(evJson)
		db.DB.Save(&stats)
	}

	c.JSON(http.StatusOK, gin.H{"message": "appeal_submitted"})
}

// HandleVote uses deterministic HMAC hashing for O(1) lookup and batch updates.
// Vote deduplication uses IP + fingerprint hash, with a global per-IP daily cap.
func HandleVote(c *gin.Context) {
	var req struct {
		TargetPhone      string `json:"target_phone"`
		TargetPhoneLocal string `json:"target_phone_local"`
		Side             string `json:"side"`
		Fingerprint      string `json:"fingerprint"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_vote_payload"})
		return
	}

	hashes := collectHashes(req.TargetPhone, req.TargetPhoneLocal)
	canonicalHash := findCanonicalHash(hashes, req.TargetPhone)

	clientIP := c.ClientIP()

	if middleware.RedisClient != nil {
		ctx := c.Request.Context()

		// 1. Per-IP daily vote cap (max 20 votes per day across all targets)
		dailyKey := "vote_daily:" + clientIP
		dailyCount, _ := middleware.RedisClient.Incr(ctx, dailyKey).Result()
		if dailyCount == 1 {
			middleware.RedisClient.Expire(ctx, dailyKey, 24*time.Hour)
		}
		if dailyCount > 20 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "vote_limit_exceeded"})
			return
		}

		// 2. IP-based dedup for this specific target
		voteKeyIP := "vote:" + canonicalHash + ":" + clientIP
		if middleware.RedisClient.Exists(ctx, voteKeyIP).Val() > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "already_voted"})
			return
		}

		// 3. Fingerprint-based dedup (if provided) for this specific target
		if req.Fingerprint != "" {
			fpHash := crypto.DeterministicHash(req.Fingerprint)
			voteKeyFP := "vote_fp:" + canonicalHash + ":" + fpHash
			if middleware.RedisClient.Exists(ctx, voteKeyFP).Val() > 0 {
				c.JSON(http.StatusConflict, gin.H{"error": "already_voted"})
				return
			}
		}
	}

	var stats model.TargetStats
	if err := db.DB.Where("target_hash = ?", canonicalHash).First(&stats).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "appeal_not_found"})
		return
	}

	if req.Side == "reporter" {
		stats.ReporterVotes++
	} else if req.Side == "appeal" {
		stats.AppealVotes++
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_side_parameter"})
		return
	}
	db.DB.Save(&stats)

	// Require higher threshold for auto-cleansing: appeal needs 2x reporter votes
	// AND at least 10 total votes to prevent small-sample manipulation.
	cleansed := false
	totalVotes := stats.ReporterVotes + stats.AppealVotes
	if totalVotes >= 10 && float64(stats.AppealVotes) > float64(stats.ReporterVotes)*2.0 {
		db.DB.Model(&model.RiskRecord{}).
			Where("target_hash IN ?", hashes).
			Update("status", "cleansed_by_jury")
		db.DB.Model(&model.RiskRecord{}).
			Where("target_local_hash IN ?", hashes).
			Update("status", "cleansed_by_jury")
		cleansed = true
	}

	// Mark IP + fingerprint as voted (TTL 30 days)
	if middleware.RedisClient != nil {
		ctx := c.Request.Context()
		voteKeyIP := "vote:" + canonicalHash + ":" + clientIP
		middleware.RedisClient.Set(ctx, voteKeyIP, "1", 30*24*time.Hour)
		if req.Fingerprint != "" {
			fpHash := crypto.DeterministicHash(req.Fingerprint)
			voteKeyFP := "vote_fp:" + canonicalHash + ":" + fpHash
			middleware.RedisClient.Set(ctx, voteKeyFP, "1", 30*24*time.Hour)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "vote_recorded",
		"cleansed": cleansed,
	})
}
