package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

// ApplyEffects returns the session with a decision's side effects recorded.
func ApplyEffects(s model.Session, effects model.Effects, now time.Time) model.Session {
	if effects.ChasingSignal {
		s.ChasingSignals++
	}
	if effects.RapidFlag {
		s.RapidFlags++
		s.RapidStreak = 0
	}
	if effects.BreakUntil != nil {
		s.BreakUntil = effects.BreakUntil
	}
	if effects.EndReason != "" {
		s.End(effects.EndReason, now)
	}
	return s
}

// SettleRound returns the plan and session after a finished round, ending the session at the stop-loss.
func SettleRound(
	plan model.Plan, s model.Session, settings model.RuleSettings, stake, payout int, now time.Time,
) (model.Plan, model.Session) {
	plan.Balance += payout - stake
	s.NetLoss += stake - payout
	s.BetCount++
	s.LastStake = stake
	s.LastLost = payout < stake
	s.RapidStreak = nextRapidStreak(s, settings, now)
	s.LastBetAt = &now
	if s.NetLoss >= settings.StopLoss {
		s.End(model.ReasonStopLossReached, now)
	}
	return plan, s
}
