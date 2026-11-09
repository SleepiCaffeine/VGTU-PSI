package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

func startAllowed() model.SessionDecision {
	return model.SessionDecision{Verdict: allow()}
}

func denied(reason model.Reason) model.SessionDecision {
	return model.SessionDecision{Verdict: block(reason)}
}

func deniedUntilNextAllocation(reason model.Reason, plan model.Plan) model.SessionDecision {
	next := NextWeekStart(plan.WeekStart)
	return model.SessionDecision{Verdict: block(reason), NextAllocationAt: &next}
}

type noOpenSessionRule struct{}

func (noOpenSessionRule) Name() string { return "no_open_session" }

func (noOpenSessionRule) Evaluate(in SessionStartInput) model.SessionDecision {
	if in.OpenSession == nil {
		return startAllowed()
	}
	return denied(model.ReasonSessionAlreadyActive)
}

type noCreditsRule struct{}

func (noCreditsRule) Name() string { return "no_credits" }

func (noCreditsRule) Evaluate(in SessionStartInput) model.SessionDecision {
	if in.Plan.Balance > 0 {
		return startAllowed()
	}
	return deniedUntilNextAllocation(model.ReasonNoCredits, in.Plan)
}

type allowedHoursRule struct{}

func (allowedHoursRule) Name() string { return "allowed_hours" }

func (allowedHoursRule) Evaluate(in SessionStartInput) model.SessionDecision {
	hour := in.Now.Hour()
	if hour >= in.Settings.AllowedFromHour && hour < in.Settings.AllowedToHour {
		return startAllowed()
	}
	return denied(model.ReasonOutsideAllowedHours)
}

type weeklySessionLimitRule struct{}

func (weeklySessionLimitRule) Name() string { return "weekly_session_limit" }

func (weeklySessionLimitRule) Evaluate(in SessionStartInput) model.SessionDecision {
	if in.Week.SessionsStarted < in.Settings.WeeklySessions {
		return startAllowed()
	}
	return denied(model.ReasonWeeklySessionLimit)
}

type cooldownRule struct{}

func (cooldownRule) Name() string { return "cooldown" }

func (cooldownRule) Evaluate(in SessionStartInput) model.SessionDecision {
	if in.Week.LastSessionEndedAt == nil {
		return startAllowed()
	}
	cooldown := time.Duration(in.Settings.CooldownMinutes) * time.Minute
	if !in.Now.Before(in.Week.LastSessionEndedAt.Add(cooldown)) {
		return startAllowed()
	}
	return denied(model.ReasonCooldownActive)
}
