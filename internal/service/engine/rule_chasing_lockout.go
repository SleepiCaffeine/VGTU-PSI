package engine

import "github.com/SleepiCaffeine/VGTU-PSI/internal/model"

type chasingLockoutRule struct{}

func (chasingLockoutRule) Name() string { return "chasing_lockout" }

func (chasingLockoutRule) Evaluate(in Input, d *model.Decision) bool {
	if in.ChasingStopsThisWeek == 0 {
		return false
	}
	next := NextWeekStart(in.Plan.WeekStart)
	d.NextAllocationAt = &next
	return deny(d, model.ReasonChasingLockout)
}

type breakActiveRule struct{}

func (breakActiveRule) Name() string { return "break_active" }

func (breakActiveRule) Evaluate(in Input, d *model.Decision) bool {
	if in.LatestBreakUntil == nil || !in.Now.Before(*in.LatestBreakUntil) {
		return false
	}
	return deny(d, model.ReasonBreakActive)
}
