package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

// SessionExpired reports whether the session has outlived its time limit.
func SessionExpired(s model.Session, settings model.RuleSettings, now time.Time) bool {
	limit := time.Duration(settings.SessionMinutes) * time.Minute
	return !now.Before(s.StartedAt.Add(limit))
}

type sessionActiveRule struct{}

func (sessionActiveRule) Name() string { return "session_active" }

func (sessionActiveRule) Evaluate(in Input, d *model.Decision) bool {
	if in.Session == nil || in.Session.Ended() {
		return reject(d, model.ReasonSessionEnded)
	}
	if !SessionExpired(*in.Session, in.Settings, in.Now) {
		return false
	}
	d.Effects.EndReason = model.ReasonSessionTimeLimit
	return reject(d, model.ReasonSessionTimeLimit)
}
