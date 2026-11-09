package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

// Input is everything a rule may consider. Rules perform no I/O.
type Input struct {
	Now                    time.Time
	Settings               model.RuleSettings
	Plan                   model.Plan
	Session                *model.Session
	SessionsThisWeek       int
	LastSessionEndedAt     *time.Time
	ChasingStopsThisWeek   int
	LatestBreakUntil       *time.Time
	ChasingSignalsThisWeek int

	RequestedStake        int
	HouseEdge             float64
	ConfirmationSubmitted bool
	Confirmation          *model.Confirmation
	ConfirmationText      string
}

// Rule is one step of a decision pipeline. It returns true to stop the pipeline.
type Rule interface {
	Name() string
	Evaluate(in Input, d *model.Decision) (stop bool)
}
