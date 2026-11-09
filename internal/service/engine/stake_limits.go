package engine

import "github.com/SleepiCaffeine/VGTU-PSI/internal/model"

// limitStake returns the stake that may be played and why. A zero stake means none may be.
func limitStake(in BetInput) (int, model.Reason) {
	if in.RequestedStake <= 0 {
		return 0, model.ReasonInvalidStake
	}
	allowed, capReason := tightestLimit(in)
	if allowed <= 0 {
		return 0, model.ReasonNoStakeAvailable
	}
	if in.RequestedStake <= allowed {
		return in.RequestedStake, model.ReasonRulesPassed
	}
	return allowed, capReason
}

func tightestLimit(in BetInput) (int, model.Reason) {
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
