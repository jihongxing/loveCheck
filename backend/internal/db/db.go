package db

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"lovecheck/internal/model"
	"lovecheck/pkg/logger"
)

// DB is the global PostgreSQL database instance.
var DB *gorm.DB

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// InitDB initializes a PostgreSQL connection with connection pooling and
// creates optimized Hash / Partial indexes after auto-migration.
func InitDB() {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		getEnv("PG_HOST", "localhost"),
		getEnv("PG_PORT", "5432"),
		getEnv("PG_USER", "lovecheck"),
		getEnv("PG_PASSWORD", "lovecheck_pwd"),
		getEnv("PG_DBNAME", "lovecheck"),
		getEnv("PG_SSLMODE", "disable"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to connect to PostgreSQL")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Failed to get underlying *sql.DB")
	}
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	err = DB.AutoMigrate(
		&model.RiskRecord{},
		&model.TargetStats{},
		&model.ActivationCode{},
		&model.PurchasePlatform{},
		&model.PushSubscription{},
		&model.PaymentOrder{},
	)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Auto-migration failed")
	}

	fixLegacyVoteDefaults()

	createOptimizedIndexes()

	logger.Log.Info().Msg("PostgreSQL connected, pooled, migrated and indexed")
}

// fixLegacyVoteDefaults resets rows that still carry the old GORM default vote
// pair (18 / 4) from an earlier schema. Real votes start at 0 and increment
// via HandleVote; the 18+4 pair was display-only seed data, not user votes.
func fixLegacyVoteDefaults() {
	res := DB.Exec(`
		UPDATE target_stats
		SET reporter_votes = 0, appeal_votes = 0
		WHERE reporter_votes = 18 AND appeal_votes = 4
	`)
	if res.Error != nil {
		logger.Log.Warn().Err(res.Error).Msg("Legacy vote default cleanup skipped")
		return
	}
	if res.RowsAffected > 0 {
		logger.Log.Info().Int64("rows", res.RowsAffected).Msg("Reset legacy default jury votes (18/4) to 0/0")
	}
}

// createOptimizedIndexes builds Hash indexes and Partial indexes that GORM
// tags cannot express. These are idempotent (IF NOT EXISTS).
func createOptimizedIndexes() {
	indexes := []string{
		// Hash indexes for O(1) exact-match lookups on 64-char HMAC hex strings
		`CREATE INDEX IF NOT EXISTS idx_hash_target
		 ON risk_records USING HASH (target_hash)`,

		`CREATE INDEX IF NOT EXISTS idx_hash_target_local
		 ON risk_records USING HASH (target_local_hash)`,

		// Partial B-Tree indexes: only active records enter the index, keeping it compact
		`CREATE INDEX IF NOT EXISTS idx_active_target_hash
		 ON risk_records (target_hash) WHERE status = 'active'`,

		`CREATE INDEX IF NOT EXISTS idx_active_target_local_hash
		 ON risk_records (target_local_hash) WHERE status = 'active'`,

		// TargetStats fast lookup
		`CREATE INDEX IF NOT EXISTS idx_target_stats_hash
		 ON target_stats USING HASH (target_hash)`,

		// Activation code lookups
		`CREATE INDEX IF NOT EXISTS idx_activation_target_hash
		 ON activation_codes USING HASH (target_hash)`,

		// GIN index on JSONB tags for containment queries (@>)
		`CREATE INDEX IF NOT EXISTS idx_risk_records_tags
		 ON risk_records USING GIN (tags)`,

		// Payment order lookups by target_hash
		`CREATE INDEX IF NOT EXISTS idx_payment_target_hash
		 ON payment_orders USING HASH (target_hash)`,
	}

	for _, ddl := range indexes {
		if err := DB.Exec(ddl).Error; err != nil {
			logger.Log.Warn().Err(err).Msg("Index creation warning (may already exist)")
		}
	}
}
