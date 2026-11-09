package engine

import "github.com/SleepiCaffeine/VGTU-PSI/internal/model"

const maxFrictionLevel = 4

// FrictionLevel grows with plan weeks, bets placed this session and chasing signals.
func FrictionLevel(plan model.Plan, settings model.RuleSettings, session *model.Session) int {
	level := (plan.WeekNumber - 1) / settings.WeeksPerFrictionLevel
	if session != nil {
		level += session.BetCount/settings.BetsPerFrictionLevel + session.ChasingSignals
	}
	return min(level, maxFrictionLevel)
}
