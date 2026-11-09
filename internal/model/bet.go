package model

import "time"

type Bet struct {
	ID             int64     `gorm:"primaryKey"`
	TenantID       int64     `gorm:"not null;index"`
	SessionID      int64     `gorm:"not null;index"`
	Game           string    `gorm:"size:40;not null"`
	Selection      string    `gorm:"size:40;not null"`
	RequestedStake int       `gorm:"not null"`
	Stake          int       `gorm:"not null"`
	Payout         int       `gorm:"not null"`
	Outcome        string    `gorm:"size:40;not null"`
	CreatedAt      time.Time `gorm:"not null"`
}
