package engine

import "github.com/SleepiCaffeine/VGTU-PSI/internal/model"

const maxFrictionLevel = 4

// FrictionLevel grows with plan weeks, this week's chasing signals, bets placed this session and rapid-betting flags.
func FrictionLevel(plan model.Plan, settings model.RuleSettings, session *model.Session, weeklySignals int) int {
	level := (plan.WeekNumber-1)/settings.WeeksPerFrictionLevel + weeklySignals
	if session != nil {
		level += session.BetCount/settings.BetsPerFrictionLevel + session.RapidFlags
	}
	return min(level, maxFrictionLevel)
}
