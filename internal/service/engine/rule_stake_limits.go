package engine

import "github.com/SleepiCaffeine/VGTU-PSI/internal/model"

type stakeLimitsRule struct{}

func (stakeLimitsRule) Name() string { return "stake_limits" }

func (stakeLimitsRule) Evaluate(in Input, d *model.Decision) bool {
	if in.RequestedStake <= 0 {
		return reject(d, model.ReasonInvalidStake)
	}
	allowed, capReason := tightestLimit(in)
	if allowed <= 0 {
		return reject(d, model.ReasonNoStakeAvailable)
	}
	if in.RequestedStake <= allowed {
		return false
	}
	d.Outcome = model.OutcomeAcceptedCapped
	d.Reason = capReason
	d.Stake = allowed
	return false
}

func tightestLimit(in Input) (int, model.Reason) {
	allowed := in.Settings.MaxStake
	reason := model.ReasonMaxStakeCap
	if in.Plan.Balance < allowed {
		allowed = in.Plan.Balance
		reason = model.ReasonBalanceCap
	}
	remainingLoss := in.Settings.StopLoss - in.Session.NetLoss
	if remainingLoss < allowed {
		allowed = remainingLoss
		reason = model.ReasonStopLossCap
	}
	return allowed, reason
}
