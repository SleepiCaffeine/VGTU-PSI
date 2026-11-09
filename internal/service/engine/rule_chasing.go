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

func (forcedBreakRule) Evaluate(in Input, d *model.Decision) bool {
	if in.Session.BreakUntil == nil || !in.Now.Before(*in.Session.BreakUntil) {
		return false
	}
	return reject(d, model.ReasonChasingBreak)
}

type chasingRule struct{}

func (chasingRule) Name() string { return "chasing" }

func (chasingRule) Evaluate(in Input, d *model.Decision) bool {
	if !isChasing(*in.Session, in.RequestedStake) {
		return false
	}
	d.Effects.ChasingSignal = true
	if in.Session.ChasingSignals+1 >= chasingSignalLimit {
		d.Effects.EndReason = model.ReasonChasingLimit
		d.Effects.NotifySupervisor = true
		return reject(d, model.ReasonChasingLimit)
	}
	breakUntil := in.Now.Add(time.Duration(in.Settings.ChasingBreakSeconds) * time.Second)
	d.Effects.BreakUntil = &breakUntil
	return false
}

func isChasing(s model.Session, requestedStake int) bool {
	return s.LastLost && requestedStake >= s.LastStake*chasingStakeMultiplier
}
