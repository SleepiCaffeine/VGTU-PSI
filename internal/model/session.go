package model

import "time"

type Session struct {
	ID             int64     `gorm:"primaryKey"`
	TenantID       int64     `gorm:"not null;index:idx_sessions_tenant_started,priority:1"`
	StartedAt      time.Time `gorm:"not null;index:idx_sessions_tenant_started,priority:2"`
	EndedAt        *time.Time
	EndReason      Reason `gorm:"size:40"`
	NetLoss        int    `gorm:"not null"`
	BetCount       int    `gorm:"not null"`
	ChasingSignals int    `gorm:"not null"`
	BreakUntil     *time.Time
	LastStake      int  `gorm:"not null"`
	LastLost       bool `gorm:"not null"`
}

func (s Session) Ended() bool {
	return s.EndedAt != nil
}

// End closes the session once; the first reason recorded is kept.
func (s *Session) End(reason Reason, at time.Time) {
	if s.Ended() {
		return
	}
	s.EndedAt = &at
	s.EndReason = reason
}
