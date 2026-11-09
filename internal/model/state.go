package model

import "time"

// State is the view of a tenant's progress shared by both pair members.
type State struct {
	Plan                   Plan
	ProgressPercent        int
	FrictionLevel          int
	ChasingSignalsThisWeek int
	ActiveSession          *Session
	NextAllocationAt       time.Time
}

// RoundResult is what a game produced for an accepted bet.
type RoundResult struct {
	Outcome string
	Payout  int
}

// BetResult is a decision together with the state it left behind.
type BetResult struct {
	Decision BetDecision
	Round    *RoundResult
	Balance  int
	Session  Session
}

// SessionStartResult carries a decision and, when allowed, the new session.
type SessionStartResult struct {
	Decision SessionDecision
	Session  *Session
}
