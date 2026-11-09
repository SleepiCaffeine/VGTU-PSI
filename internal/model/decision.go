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
	ReasonChasingLimit         Reason = "CHASING_LIMIT"
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

// ConfirmationRequest tells the player what must be acknowledged before the bet.
type ConfirmationRequest struct {
	ID          int64
	Kind        ConfirmationKind
	Phrase      string
	WaitSeconds int
}

// Effects are the state changes a decision causes besides moving points.
type Effects struct {
	ChasingSignal    bool
	BreakUntil       *time.Time
	EndReason        Reason
	NotifySupervisor bool
}

// Decision is the rule engine's answer to a session start or a bet.
type Decision struct {
	Outcome          Outcome
	Reason           Reason
	RequestedStake   int
	Stake            int
	FrictionLevel    int
	Confirmation     *ConfirmationRequest
	NextAllocationAt *time.Time
	ConfirmationUsed bool
	Effects          Effects
}

func (d Decision) Accepted() bool {
	return d.Outcome == OutcomeAccepted || d.Outcome == OutcomeAcceptedCapped
}
