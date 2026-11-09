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

func (sessionActiveRule) Evaluate(in BetInput) model.Verdict {
	if in.Session.Ended() {
		return block(model.ReasonSessionEnded)
	}
	if !SessionExpired(in.Session, in.Settings, in.Now) {
		return allow()
	}
	return blockWith(model.ReasonSessionTimeLimit, model.Effects{EndReason: model.ReasonSessionTimeLimit})
}
