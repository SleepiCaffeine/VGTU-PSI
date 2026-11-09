package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

// SessionStartInput is everything a session-start rule may consider.
type SessionStartInput struct {
	Now         time.Time
	Settings    model.RuleSettings
	Plan        model.Plan
	OpenSession *model.Session
	Week        model.WeekActivity
}

// ConfirmationAttempt is an acknowledgement the player submitted with a bet.
// Record is nil when the submitted ID matches nothing in the tenant.
type ConfirmationAttempt struct {
	Record *model.Confirmation
	Text   string
}

// BetInput is everything a bet decision may consider. Attempt is nil when none was submitted.
type BetInput struct {
	Now            time.Time
	Settings       model.RuleSettings
	Plan           model.Plan
	Session        model.Session
	Week           model.WeekActivity
	RequestedStake int
	HouseEdge      float64
	Attempt        *ConfirmationAttempt
}

// SessionRule is one step of the session-start pipeline.
type SessionRule interface {
	Name() string
	Evaluate(in SessionStartInput) model.SessionDecision
}

// BetRule is a gate of the bet pipeline. A verdict that is not allowed stops the bet.
type BetRule interface {
	Name() string
	Evaluate(in BetInput) model.Verdict
}

func allow() model.Verdict {
	return model.Verdict{Allowed: true, Reason: model.ReasonRulesPassed}
}

func allowWith(effects model.Effects) model.Verdict {
	return model.Verdict{Allowed: true, Reason: model.ReasonRulesPassed, Effects: effects}
}

func block(reason model.Reason) model.Verdict {
	return model.Verdict{Reason: reason}
}

func blockWith(reason model.Reason, effects model.Effects) model.Verdict {
	return model.Verdict{Reason: reason, Effects: effects}
}
