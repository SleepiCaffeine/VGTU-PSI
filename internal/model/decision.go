package model

import "time"

type Outcome string

const (
	OutcomeAccepted             Outcome = "ACCEPTED"
	OutcomeAcceptedCapped       Outcome = "ACCEPTED_CAPPED"
	OutcomeRejected             Outcome = "REJECTED"
	OutcomeDenied               Outcome = "DENIED"
	OutcomeConfirmationRequired Outcome = "CONFIRMATION_REQUIRED"
	OutcomeConfirmationFailed   Outcome = "CONFIRMATION_FAILED"
)

type Reason string

const (
	ReasonRulesPassed          Reason = "RULES_PASSED"
	ReasonSessionAlreadyActive Reason = "SESSION_ALREADY_ACTIVE"
	ReasonNoCredits            Reason = "NO_CREDITS"
	ReasonOutsideAllowedHours  Reason = "OUTSIDE_ALLOWED_HOURS"
	ReasonWeeklySessionLimit   Reason = "WEEKLY_SESSION_LIMIT"
	ReasonCooldownActive       Reason = "COOLDOWN_ACTIVE"
	ReasonSessionEnded         Reason = "SESSION_ENDED"
	ReasonSessionTimeLimit     Reason = "SESSION_TIME_LIMIT"
	ReasonUserEnded            Reason = "USER_ENDED"
	ReasonChasingBreak         Reason = "CHASING_BREAK"
	ReasonRapidBetting         Reason = "RAPID_BETTING"
	ReasonChasingSignal        Reason = "CHASING_SIGNAL"
	ReasonChasingLimit         Reason = "CHASING_LIMIT"
	ReasonChasingLockout       Reason = "CHASING_LOCKOUT"
	ReasonBreakActive          Reason = "BREAK_ACTIVE"
	ReasonInvalidStake         Reason = "INVALID_STAKE"
	ReasonNoStakeAvailable     Reason = "NO_STAKE_AVAILABLE"
	ReasonStopLossCap          Reason = "STOP_LOSS_CAP"
	ReasonMaxStakeCap          Reason = "MAX_STAKE_CAP"
	ReasonBalanceCap           Reason = "BALANCE_CAP"
	ReasonStopLossReached      Reason = "STOP_LOSS_REACHED"
	ReasonConfirmationNeeded   Reason = "CONFIRMATION_NEEDED"
	ReasonConfirmationInvalid  Reason = "CONFIRMATION_INVALID"
	ReasonConfirmationTooEarly Reason = "CONFIRMATION_TOO_EARLY"
	ReasonConfirmationMismatch Reason = "CONFIRMATION_TEXT_MISMATCH"
)

// Effects are the state changes a verdict causes besides moving points.
type Effects struct {
	ChasingSignal bool
	RapidFlag     bool
	BreakUntil    *time.Time
	EndReason     Reason
	AlertReason   Reason
}

// Merge combines two sets of effects; values set in other take precedence.
func (e Effects) Merge(other Effects) Effects {
	merged := e
	merged.ChasingSignal = e.ChasingSignal || other.ChasingSignal
	merged.RapidFlag = e.RapidFlag || other.RapidFlag
	if other.BreakUntil != nil {
		merged.BreakUntil = other.BreakUntil
	}
	if other.EndReason != "" {
		merged.EndReason = other.EndReason
	}
	if other.AlertReason != "" {
		merged.AlertReason = other.AlertReason
	}
	return merged
}

// Verdict says whether something may proceed, why, and what follows from it.
type Verdict struct {
	Allowed bool
	Reason  Reason
	Effects Effects
}

// ConfirmationRequest tells the player what must be acknowledged before the bet.
type ConfirmationRequest struct {
	ID          int64
	Kind        ConfirmationKind
	Phrase      string
	WaitSeconds int
}

// BetDecision is the engine's answer to a bet.
type BetDecision struct {
	Outcome      Outcome
	Verdict      Verdict
	Stake        int
	Confirmation *ConfirmationRequest
}

func (d BetDecision) Accepted() bool {
	return d.Verdict.Allowed
}

// WithConfirmationID returns a copy whose confirmation request carries the stored ID.
func (d BetDecision) WithConfirmationID(id int64) BetDecision {
	request := *d.Confirmation
	request.ID = id
	d.Confirmation = &request
	return d
}

// SessionDecision is the engine's answer to starting a session.
type SessionDecision struct {
	Verdict          Verdict
	NextAllocationAt *time.Time
}

func (d SessionDecision) Accepted() bool {
	return d.Verdict.Allowed
}

func (d SessionDecision) Outcome() Outcome {
	if d.Accepted() {
		return OutcomeAccepted
	}
	return OutcomeDenied
}
