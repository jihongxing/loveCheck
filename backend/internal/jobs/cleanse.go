package jobs

import (
	"time"

	"gorm.io/gorm"

	"lovecheck/internal/model"
	"lovecheck/pkg/logger"
)

// StartCleanseScheduler runs the jury-based cleansing check on a fixed interval.
// If AppealVotes > ReporterVotes * 1.5, records are marked "cleansed_by_jury".
// If a previously cleansed target's votes shift back (reporters regain majority),
// records are restored to "active".
func StartCleanseScheduler(database *gorm.DB, interval time.Duration) {
	go func() {
		logger.Log.Info().Str("interval", interval.String()).Msg("Cleanse scheduler started")

		// Run once at startup, then on interval
		runCleanseCheck(database)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			runCleanseCheck(database)
		}
	}()
}

func runCleanseCheck(database *gorm.DB) {
	logger.Log.Info().Msg("Running periodic cleanse check")

	var allStats []model.TargetStats
	database.Find(&allStats)

	cleansedCount := 0
	restoredCount := 0

	for _, stats := range allStats {
		if stats.ReporterVotes+stats.AppealVotes == 0 {
			continue
		}

		shouldCleanse := float64(stats.AppealVotes) > float64(stats.ReporterVotes)*1.5
		hash := stats.TargetHash

		if shouldCleanse {
			result := database.Model(&model.RiskRecord{}).
				Where("(target_hash = ? OR target_local_hash = ?) AND status = ?", hash, hash, "active").
				Update("status", "cleansed_by_jury")
			if result.RowsAffected > 0 {
				cleansedCount += int(result.RowsAffected)
			}
		} else {
			result := database.Model(&model.RiskRecord{}).
				Where("(target_hash = ? OR target_local_hash = ?) AND status = ?", hash, hash, "cleansed_by_jury").
				Update("status", "active")
			if result.RowsAffected > 0 {
				restoredCount += int(result.RowsAffected)
			}
		}
	}

	logger.Log.Info().
		Int("total_targets", len(allStats)).
		Int("cleansed_records", cleansedCount).
		Int("restored_records", restoredCount).
		Msg("Cleanse check completed")
}
