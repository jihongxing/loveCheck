package bloom

import (
	"sync"

	bloomfilter "github.com/bits-and-blooms/bloom/v3"
	"gorm.io/gorm"

	"lovecheck/internal/model"
	"lovecheck/pkg/logger"
)

var (
	filter *bloomfilter.BloomFilter
	mu     sync.RWMutex
)

// Init creates the bloom filter and pre-loads all existing hashes from the DB.
// expectedN = estimated max records; fpRate = acceptable false positive rate.
// For 10M records at 0.01% FP rate, memory usage is ~23 MB.
func Init(database *gorm.DB, expectedN uint, fpRate float64) {
	mu.Lock()
	defer mu.Unlock()

	filter = bloomfilter.NewWithEstimates(expectedN, fpRate)

	var hashes []struct {
		TargetHash      string
		TargetLocalHash string
	}
	database.Model(&model.RiskRecord{}).
		Select("target_hash, target_local_hash").
		Where("status = ?", "active").
		FindInBatches(&hashes, 5000, func(tx *gorm.DB, batch int) error {
			for _, h := range hashes {
				if h.TargetHash != "" {
					filter.AddString(h.TargetHash)
				}
				if h.TargetLocalHash != "" {
					filter.AddString(h.TargetLocalHash)
				}
			}
			return nil
		})

	logger.Log.Info().Uint("capacity", expectedN).Float64("fp_rate_pct", fpRate*100).Msg("Bloom filter initialized")
}

// Add inserts a hash into the bloom filter (call after inserting a new record).
func Add(hash string) {
	if hash == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	filter.AddString(hash)
}

// MayExist returns false if the hash definitely does NOT exist,
// or true if the hash MIGHT exist (requires DB confirmation).
func MayExist(hash string) bool {
	mu.RLock()
	defer mu.RUnlock()
	return filter.TestString(hash)
}

// MayExistAny returns true if any of the provided hashes might exist.
func MayExistAny(hashes []string) bool {
	mu.RLock()
	defer mu.RUnlock()
	for _, h := range hashes {
		if filter.TestString(h) {
			return true
		}
	}
	return false
}
