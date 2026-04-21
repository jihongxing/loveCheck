package model

import "time"

// AccessGrant scopes unlock access for a single client token to a single target.
// This prevents one user's unlock from becoming globally visible to everyone.
type AccessGrant struct {
	ID              uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	TargetHash      string     `gorm:"type:char(64);not null;index" json:"-"`
	ClientTokenHash string     `gorm:"type:char(64);not null;index" json:"-"`
	Source          string     `gorm:"type:varchar(30);not null" json:"source"`
	SourceRef       string     `gorm:"type:varchar(64)" json:"source_ref"`
	Status          string     `gorm:"type:varchar(20);default:'active';index" json:"status"`
	CreatedAt       time.Time  `gorm:"autoCreateTime;index" json:"created_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
}
