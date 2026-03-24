package model

import (
	"time"
)

// RiskRecord represents a report in the PostgreSQL database.
// Indexes are created via raw SQL in db.go (Hash + Partial indexes).
type RiskRecord struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetHash      string    `gorm:"type:char(64);not null" json:"-"`
	TargetLocalHash string    `gorm:"type:char(64)" json:"-"`
	DisplayName     string    `gorm:"type:varchar(100)" json:"display_name"`
	LocationCity    string    `gorm:"type:varchar(200)" json:"location_city"`
	Description     string    `gorm:"type:text" json:"description"`
	RiskLevel       int       `gorm:"default:1" json:"risk_level"`
	Tags            string    `gorm:"type:jsonb;default:'[]'" json:"tags"`
	EvidenceMaskURL string    `gorm:"type:text" json:"evidence_mask_url"`
	Status          string    `gorm:"type:varchar(30);default:'active';index" json:"status"`
	ReporterHash    string    `gorm:"type:char(64);not null" json:"-"`
	// Virtual identity & trust signals (no phone fragments — avoids reprisal risk).
	ReporterDisplayName string `gorm:"type:varchar(80)" json:"reporter_display_name"`
	VerificationLevel   int    `gorm:"default:1" json:"verification_level"` // 1=phone path, 2=reserved ID/KYC, 3=evidence attached
	ReporterCity        string `gorm:"type:varchar(100)" json:"reporter_city"` // trusted proxy geo label, optional
	CreatedAt           time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

// ActivationCode represents a pre-generated unlock code sold via third-party card platforms.
type ActivationCode struct {
	ID          uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	Code        string     `gorm:"type:varchar(20);uniqueIndex;not null" json:"code"`
	Status      string     `gorm:"type:varchar(10);default:'unused';index" json:"status"`
	TargetHash  string     `gorm:"type:char(64)" json:"-"`
	ActivatedIP string     `gorm:"type:varchar(45)" json:"-"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
	ActivatedAt *time.Time `json:"activated_at"`
}

// PurchasePlatform represents a configurable external store where users buy activation codes.
type PurchasePlatform struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	URL       string    `gorm:"type:varchar(500);not null" json:"url"`
	Icon      string    `gorm:"type:varchar(500)" json:"icon"`
	Region    string    `gorm:"type:varchar(50)" json:"region"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	Active    bool      `gorm:"default:true" json:"active"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// PushSubscription stores browser Web Push subscription info for watch notifications.
type PushSubscription struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Endpoint   string    `gorm:"type:text;not null;uniqueIndex" json:"endpoint"`
	KeyAuth    string    `gorm:"type:varchar(200);not null" json:"-"`
	KeyP256dh  string    `gorm:"type:varchar(200);not null" json:"-"`
	TargetHash string    `gorm:"type:char(64);not null;index" json:"-"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

// PaymentOrder tracks payment orders (WeChat/Alipay via Xunhupay, PayPal) for unlocking query results.
type PaymentOrder struct {
	ID            uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderNo       string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"order_no"`
	TargetHash    string     `gorm:"type:char(64);not null;index" json:"-"`
	Provider      string     `gorm:"type:varchar(10);default:'wechat'" json:"provider"`
	Amount        string     `gorm:"type:varchar(10);not null" json:"amount"`
	Currency      string     `gorm:"type:varchar(5);default:'CNY'" json:"currency"`
	Status        string     `gorm:"type:varchar(10);default:'pending';index" json:"status"`
	XunhuOrderID  string     `gorm:"type:varchar(64)" json:"-"`
	PayPalOrderID string     `gorm:"type:varchar(64)" json:"-"`
	TransactionID string     `gorm:"type:varchar(64)" json:"-"`
	ClientIP      string     `gorm:"type:varchar(45)" json:"-"`
	CreatedAt     time.Time  `gorm:"autoCreateTime" json:"created_at"`
	PaidAt        *time.Time `json:"paid_at"`
}

// TargetStats represents the global consensus state (appeal info & votes) for a unique target.
type TargetStats struct {
	TargetHash     string    `gorm:"primaryKey;type:char(64)" json:"target_hash"`
	AppealReason   string    `gorm:"type:text" json:"appeal_reason"`
	AppealEvidence string    `gorm:"type:text" json:"appeal_evidence"`
	ReporterVotes  int       `gorm:"default:0" json:"reporter_votes"`
	AppealVotes    int       `gorm:"default:0" json:"appeal_votes"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}
