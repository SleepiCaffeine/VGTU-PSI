package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type rapidBettingRule struct{}

func (rapidBettingRule) Name() string { return "rapid_betting" }

func (rapidBettingRule) Evaluate(in Input, d *model.Decision) bool {
	limit := in.Settings.RapidBetLimit
	if limit == 0 || nextRapidStreak(*in.Session, in.Settings, in.Now) < limit {
		return false
	}
	breakUntil := in.Now.Add(time.Duration(in.Settings.ChasingBreakSeconds) * time.Second)
	d.Effects.RapidFlag = true
	d.Effects.BreakUntil = &breakUntil
	d.Effects.AlertReason = model.ReasonRapidBetting
	return reject(d, model.ReasonRapidBetting)
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
