package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

const (
	chasingStakeMultiplier = 2
	chasingSignalLimit     = 3
)

type forcedBreakRule struct{}

func (forcedBreakRule) Name() string { return "forced_break" }

func (forcedBreakRule) Evaluate(in BetInput) model.Verdict {
	if in.Session.BreakUntil == nil || !in.Now.Before(*in.Session.BreakUntil) {
		return allow()
	}
	return block(model.ReasonChasingBreak)
}

type chasingRule struct{}

func (chasingRule) Name() string { return "chasing" }

func (chasingRule) Evaluate(in BetInput) model.Verdict {
	if !isChasing(in.Session, in.RequestedStake) {
		return allow()
	}
	if in.Week.ChasingSignals+1 >= chasingSignalLimit {
		return blockWith(model.ReasonChasingLimit, model.Effects{
			ChasingSignal: true,
			EndReason:     model.ReasonChasingLimit,
			AlertReason:   model.ReasonChasingLimit,
		})
	}
	breakUntil := in.Now.Add(time.Duration(in.Settings.ChasingBreakSeconds) * time.Second)
	return allowWith(model.Effects{
		ChasingSignal: true,
		BreakUntil:    &breakUntil,
		AlertReason:   model.ReasonChasingSignal,
	})
}

func isChasing(s model.Session, requestedStake int) bool {
	return s.LastLost && requestedStake >= s.LastStake*chasingStakeMultiplier
}
