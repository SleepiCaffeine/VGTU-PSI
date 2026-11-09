package model

// RuleSettings holds the limits a supervisor sets for a tenant.
type RuleSettings struct {
	ID                    int64 `gorm:"primaryKey"`
	TenantID              int64 `gorm:"not null;uniqueIndex"`
	MaxStake              int   `gorm:"not null"`
	StopLoss              int   `gorm:"not null"`
	SessionMinutes        int   `gorm:"not null"`
	WeeklySessions        int   `gorm:"not null"`
	CooldownMinutes       int   `gorm:"not null"`
	AllowedFromHour       int   `gorm:"not null"`
	AllowedToHour         int   `gorm:"not null"`
	ChasingBreakSeconds   int   `gorm:"not null"`
	ConfirmEveryBets      int   `gorm:"not null"`
	FrictionWaitSeconds   int   `gorm:"not null"`
	WeeksPerFrictionLevel int   `gorm:"not null"`
	BetsPerFrictionLevel  int   `gorm:"not null"`
	RapidBetSeconds       int   `gorm:"not null"`
	RapidBetLimit         int   `gorm:"not null"`
}
