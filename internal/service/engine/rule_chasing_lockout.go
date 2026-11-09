package engine

import "github.com/SleepiCaffeine/VGTU-PSI/internal/model"

type chasingLockoutRule struct{}

func (chasingLockoutRule) Name() string { return "chasing_lockout" }

func (chasingLockoutRule) Evaluate(in SessionStartInput) model.SessionDecision {
	if in.Week.ChasingStops == 0 {
		return startAllowed()
	}
	return deniedUntilNextAllocation(model.ReasonChasingLockout, in.Plan)
}

type breakActiveRule struct{}

func (breakActiveRule) Name() string { return "break_active" }

func (breakActiveRule) Evaluate(in SessionStartInput) model.SessionDecision {
	if in.Week.LatestBreakUntil == nil || !in.Now.Before(*in.Week.LatestBreakUntil) {
		return startAllowed()
	}
	return denied(model.ReasonBreakActive)
}
