// Package engine decides whether a session may start or a bet may be placed.
package engine

import "github.com/SleepiCaffeine/VGTU-PSI/internal/model"

type Engine struct {
	sessionRules []Rule
	betRules     []Rule
}

func New(sessionRules, betRules []Rule) *Engine {
	return &Engine{sessionRules: sessionRules, betRules: betRules}
}

func (e *Engine) DecideSessionStart(in Input) model.Decision {
	d := model.Decision{Outcome: model.OutcomeAccepted, Reason: model.ReasonRulesPassed}
	run(e.sessionRules, in, &d)
	return d
}

func (e *Engine) DecideBet(in Input) model.Decision {
	d := model.Decision{
		Outcome:        model.OutcomeAccepted,
		Reason:         model.ReasonRulesPassed,
		RequestedStake: in.RequestedStake,
		Stake:          in.RequestedStake,
		FrictionLevel:  FrictionLevel(in.Plan, in.Settings, in.Session, in.ChasingSignalsThisWeek),
	}
	run(e.betRules, in, &d)
	return d
}

func run(rules []Rule, in Input, d *model.Decision) {
	for _, rule := range rules {
		if rule.Evaluate(in, d) {
			return
		}
	}
}

func reject(d *model.Decision, reason model.Reason) bool {
	d.Outcome = model.OutcomeRejected
	d.Reason = reason
	d.Stake = 0
	return true
}

func deny(d *model.Decision, reason model.Reason) bool {
	d.Outcome = model.OutcomeDenied
	d.Reason = reason
	return true
}
