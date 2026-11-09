package model

import "time"

// WeekActivity summarizes a tenant's session history for rule decisions.
// The counts cover one plan week; the two times are the latest across all weeks.
type WeekActivity struct {
	SessionsStarted    int
	ChasingSignals     int
	ChasingStops       int
	LastSessionEndedAt *time.Time
	LatestBreakUntil   *time.Time
}
