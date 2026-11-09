package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

// ApplyEffects records a decision's side effects on the session.
func ApplyEffects(s *model.Session, effects model.Effects, now time.Time) {
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
}

// SettleRound moves points for a finished round and ends the session at the stop-loss.
func SettleRound(plan *model.Plan, s *model.Session, settings model.RuleSettings, stake, payout int, now time.Time) {
	plan.Balance += payout - stake
	s.NetLoss += stake - payout
	s.BetCount++
	s.LastStake = stake
	s.LastLost = payout < stake
	s.RapidStreak = nextRapidStreak(*s, settings, now)
	s.LastBetAt = &now
	if s.NetLoss >= settings.StopLoss {
		s.End(model.ReasonStopLossReached, now)
	}
}
