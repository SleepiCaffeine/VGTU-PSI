package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type rapidBettingRule struct{}

func (rapidBettingRule) Name() string { return "rapid_betting" }

func (rapidBettingRule) Evaluate(in BetInput) model.Verdict {
	limit := in.Settings.RapidBetLimit
	if limit == 0 || nextRapidStreak(in.Session, in.Settings, in.Now) < limit {
		return allow()
	}
	breakUntil := in.Now.Add(time.Duration(in.Settings.ChasingBreakSeconds) * time.Second)
	return blockWith(model.ReasonRapidBetting, model.Effects{
		RapidFlag:   true,
		BreakUntil:  &breakUntil,
		AlertReason: model.ReasonRapidBetting,
	})
}

// nextRapidStreak counts consecutive bets placed too soon after the previous one.
func nextRapidStreak(s model.Session, settings model.RuleSettings, now time.Time) int {
	if s.LastBetAt == nil {
		return 0
	}
	if now.Sub(*s.LastBetAt) >= time.Duration(settings.RapidBetSeconds)*time.Second {
		return 0
	}
	return s.RapidStreak + 1
}
