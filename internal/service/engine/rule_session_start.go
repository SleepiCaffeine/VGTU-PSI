package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type noOpenSessionRule struct{}

func (noOpenSessionRule) Name() string { return "no_open_session" }

func (noOpenSessionRule) Evaluate(in Input, d *model.Decision) bool {
	if in.Session == nil {
		return false
	}
	return deny(d, model.ReasonSessionAlreadyActive)
}

type noCreditsRule struct{}

func (noCreditsRule) Name() string { return "no_credits" }

func (noCreditsRule) Evaluate(in Input, d *model.Decision) bool {
	if in.Plan.Balance > 0 {
		return false
	}
	next := NextWeekStart(in.Plan.WeekStart)
	d.NextAllocationAt = &next
	return deny(d, model.ReasonNoCredits)
}

type allowedHoursRule struct{}

func (allowedHoursRule) Name() string { return "allowed_hours" }

func (allowedHoursRule) Evaluate(in Input, d *model.Decision) bool {
	hour := in.Now.Hour()
	if hour >= in.Settings.AllowedFromHour && hour < in.Settings.AllowedToHour {
		return false
	}
	return deny(d, model.ReasonOutsideAllowedHours)
}

type weeklySessionLimitRule struct{}

func (weeklySessionLimitRule) Name() string { return "weekly_session_limit" }

func (weeklySessionLimitRule) Evaluate(in Input, d *model.Decision) bool {
	if in.SessionsThisWeek < in.Settings.WeeklySessions {
		return false
	}
	return deny(d, model.ReasonWeeklySessionLimit)
}

type cooldownRule struct{}

func (cooldownRule) Name() string { return "cooldown" }

func (cooldownRule) Evaluate(in Input, d *model.Decision) bool {
	if in.LastSessionEndedAt == nil {
		return false
	}
	cooldown := time.Duration(in.Settings.CooldownMinutes) * time.Minute
	if !in.Now.Before(in.LastSessionEndedAt.Add(cooldown)) {
		return false
	}
	return deny(d, model.ReasonCooldownActive)
}
