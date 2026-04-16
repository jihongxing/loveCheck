package model

import (
	"time"
)

// CompanyRecord represents a company abuse report in the database.
// Similar to RiskRecord but tailored for corporate entities.
type CompanyRecord struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyHash     string    `gorm:"type:char(64);not null;index" json:"-"`                    // Hash of company name + registration number
	CompanyName     string    `gorm:"type:varchar(200);not null" json:"company_name"`           // Full company name
	DisplayName     string    `gorm:"type:varchar(200)" json:"display_name"`                    // Masked company name for public display
	RegistrationNo  string    `gorm:"type:varchar(100)" json:"registration_no"`                 // Masked registration number (e.g., 91****1234)
	Industry        string    `gorm:"type:varchar(100)" json:"industry"`                        // Industry category
	LocationCity    string    `gorm:"type:varchar(200)" json:"location_city"`                   // Company location
	Description     string    `gorm:"type:text" json:"description"`                             // Detailed abuse description
	RiskLevel       int       `gorm:"default:1" json:"risk_level"`                              // 1=low, 2=medium, 3=high
	Tags            string    `gorm:"type:jsonb;default:'[]'" json:"tags"`                      // Abuse tags: ["拖欠工资", "强制加班", "PUA管理", "违法裁员"]
	EvidenceMaskURL string    `gorm:"type:text" json:"evidence_mask_url"`                       // JSON array of evidence file URLs
	Status          string    `gorm:"type:varchar(30);default:'active';index" json:"status"`    // active, hidden, cleansed_by_jury
	ReporterHash    string    `gorm:"type:char(64);not null" json:"-"`                          // Reporter identity hash
	ReporterDisplayName string `gorm:"type:varchar(80)" json:"reporter_display_name"`           // Anonymous reporter nickname
	VerificationLevel   int    `gorm:"default:1" json:"verification_level"`                     // 1=basic, 2=with contract, 3=with legal docs
	ReporterCity        string `gorm:"type:varchar(100)" json:"reporter_city"`                  // Reporter location
	EmploymentPeriod    string `gorm:"type:varchar(100)" json:"employment_period"`              // e.g., "2023.01-2024.06"
	Position            string `gorm:"type:varchar(100)" json:"position"`                       // Job position (optional, masked)
	CreatedAt           time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// CompanyStats represents the global consensus state for a company (appeal & votes).
type CompanyStats struct {
	CompanyHash    string    `gorm:"primaryKey;type:char(64)" json:"company_hash"`
	AppealReason   string    `gorm:"type:text" json:"appeal_reason"`
	AppealEvidence string    `gorm:"type:text" json:"appeal_evidence"`
	ReporterVotes  int       `gorm:"default:0" json:"reporter_votes"`  // Votes supporting reporters
	CompanyVotes   int       `gorm:"default:0" json:"company_votes"`   // Votes supporting company appeal
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// CompanyWatchlist allows users to subscribe to updates about specific companies.
type CompanyWatchlist struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	CompanyHash string    `gorm:"type:char(64);not null;index" json:"-"`
	UserHash    string    `gorm:"type:char(64);not null;index" json:"-"`  // Subscriber identity hash
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}
