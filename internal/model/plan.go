package model

import "time"

// Plan is a tenant's weekly allocation schedule and current balance.
type Plan struct {
	ID                int64     `gorm:"primaryKey"`
	TenantID          int64     `gorm:"not null;uniqueIndex"`
	InitialAllocation int       `gorm:"not null"`
	Allocation        int       `gorm:"not null"`
	Balance           int       `gorm:"not null"`
	WeekNumber        int       `gorm:"not null"`
	WeekStart         time.Time `gorm:"not null"`
}

// ProgressPercent reports how far the allocation has fallen since the plan began.
func (p Plan) ProgressPercent() int {
	if p.InitialAllocation == 0 {
		return 0
	}
	return (p.InitialAllocation - p.Allocation) * 100 / p.InitialAllocation
}
