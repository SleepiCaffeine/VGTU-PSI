package engine

import (
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

const (
	reducedAllocationPercent = 85
	allocationRounding       = 10
	daysPerWeek              = 7
)

// ReduceAllocation lowers an allocation by 15%, rounded down to tens.
func ReduceAllocation(allocation int) int {
	return allocation * reducedAllocationPercent / 100 / allocationRounding * allocationRounding
}

// WeekStartOf returns Monday 00:00 of the week containing t, in t's location.
func WeekStartOf(t time.Time) time.Time {
	daysSinceMonday := (int(t.Weekday()) + 6) % daysPerWeek
	return time.Date(t.Year(), t.Month(), t.Day()-daysSinceMonday, 0, 0, 0, 0, t.Location())
}

func NextWeekStart(weekStart time.Time) time.Time {
	return weekStart.AddDate(0, 0, daysPerWeek)
}

// AdvanceWeek returns the plan for the next week; unused points do not carry over.
func AdvanceWeek(plan model.Plan, clean bool) model.Plan {
	if clean {
		plan.Allocation = ReduceAllocation(plan.Allocation)
	}
	plan.Balance = plan.Allocation
	plan.WeekNumber++
	plan.WeekStart = NextWeekStart(plan.WeekStart)
	return plan
}
