// Package engine decides whether a session may start or a bet may be placed.
package engine

import "github.com/SleepiCaffeine/VGTU-PSI/internal/model"

type Engine struct {
	sessionRules []SessionRule
	betGates     []BetRule
}

func New(sessionRules []SessionRule, betGates []BetRule) *Engine {
	return &Engine{sessionRules: sessionRules, betGates: betGates}
}

func (e *Engine) DecideSessionStart(in SessionStartInput) model.SessionDecision {
	for _, rule := range e.sessionRules {
		decision := rule.Evaluate(in)
		if !decision.Accepted() {
			return decision
		}
	}
	return startAllowed()
}

// DecideBet runs three phases: gates, then the stake limit, then friction.
func (e *Engine) DecideBet(in BetInput) model.BetDecision {
	effects := model.Effects{}
	for _, gate := range e.betGates {
		verdict := gate.Evaluate(in)
		effects = effects.Merge(verdict.Effects)
		if !verdict.Allowed {
			return rejected(verdict.Reason, effects)
		}
	}

	stake, stakeReason := limitStake(in)
	if stake == 0 {
		return rejected(stakeReason, effects)
	}

	level := FrictionLevel(in.Plan, in.Settings, &in.Session, in.Week.ChasingSignals)
	request := requiredConfirmation(in, level, stake)
	if request == nil {
		return accepted(stake, stakeReason, effects)
	}
	if in.Attempt == nil {
		return awaitingConfirmation(stake, request)
	}
	failure := confirmationFailure(*in.Attempt, in, stake, request)
	if failure != "" {
		return confirmationFailed(stake, failure)
	}
	return accepted(stake, stakeReason, effects)
}

func rejected(reason model.Reason, effects model.Effects) model.BetDecision {
	return model.BetDecision{Outcome: model.OutcomeRejected, Verdict: blockWith(reason, effects)}
}

func accepted(stake int, reason model.Reason, effects model.Effects) model.BetDecision {
	outcome := model.OutcomeAcceptedCapped
	if reason == model.ReasonRulesPassed {
		outcome = model.OutcomeAccepted
	}
	return model.BetDecision{
		Outcome: outcome,
		Verdict: model.Verdict{Allowed: true, Reason: reason, Effects: effects},
		Stake:   stake,
	}
}

// awaitingConfirmation carries no effects so that a retried bet cannot count them twice.
func awaitingConfirmation(stake int, request *model.ConfirmationRequest) model.BetDecision {
	return model.BetDecision{
		Outcome:      model.OutcomeConfirmationRequired,
		Verdict:      block(model.ReasonConfirmationNeeded),
		Stake:        stake,
		Confirmation: request,
	}
}

func confirmationFailed(stake int, reason model.Reason) model.BetDecision {
	return model.BetDecision{Outcome: model.OutcomeConfirmationFailed, Verdict: block(reason), Stake: stake}
}
