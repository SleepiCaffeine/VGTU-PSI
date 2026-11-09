package model

import "time"

// Alert notifies the supervisor of an event in the tenant's play.
type Alert struct {
	ID        int64     `gorm:"primaryKey"`
	TenantID  int64     `gorm:"not null;index"`
	SessionID int64     `gorm:"not null"`
	Reason    Reason    `gorm:"size:40;not null"`
	CreatedAt time.Time `gorm:"not null"`
}
