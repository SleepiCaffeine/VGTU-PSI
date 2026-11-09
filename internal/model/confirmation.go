package model

import "time"

type ConfirmationKind string

const (
	ConfirmationClick ConfirmationKind = "CLICK"
	ConfirmationText  ConfirmationKind = "TEXT"
)

// Confirmation is a server-side record of an acknowledgement that covers one bet.
type Confirmation struct {
	ID        int64            `gorm:"primaryKey"`
	TenantID  int64            `gorm:"not null;index"`
	SessionID int64            `gorm:"not null;index"`
	Stake     int              `gorm:"not null"`
	Kind      ConfirmationKind `gorm:"size:10;not null"`
	Phrase    string           `gorm:"size:255;not null"`
	NotBefore time.Time        `gorm:"not null"`
	UsedAt    *time.Time
	CreatedAt time.Time `gorm:"not null"`
}
