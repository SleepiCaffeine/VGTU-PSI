package service

import (
	"context"
	"fmt"
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
)

const demoWeek = 7 * 24 * time.Hour

type PlanService struct {
	deps Dependencies
}

func NewPlanService(deps Dependencies) *PlanService {
	return &PlanService{deps: deps}
}

func (s *PlanService) State(ctx context.Context, tenantID int64) (model.State, error) {
	var state model.State
	err := s.deps.UnitOfWork.Do(ctx, func(ctx context.Context, repos Repositories) error {
		var err error
		state, err = loadState(ctx, repos, tenantID, s.deps.Clock)
		return err
	})
	return state, err
}

// AdvanceWeek moves the tenant's history one week into the past so a demo can reach a week change.
func (s *PlanService) AdvanceWeek(ctx context.Context, tenantID int64) (model.State, error) {
	var state model.State
	err := s.deps.UnitOfWork.Do(ctx, func(ctx context.Context, repos Repositories) error {
		plan, err := repos.Plans.GetForUpdate(ctx, tenantID)
		if err != nil {
			return fmt.Errorf("load plan: %w", err)
		}
		plan.WeekStart = plan.WeekStart.Add(-demoWeek)
		err = repos.Plans.Save(ctx, plan)
		if err != nil {
			return fmt.Errorf("save plan: %w", err)
		}
		err = repos.Sessions.ShiftBack(ctx, tenantID, demoWeek)
		if err != nil {
			return fmt.Errorf("shift sessions: %w", err)
		}
		state, err = loadState(ctx, repos, tenantID, s.deps.Clock)
		return err
	})
	return state, err
}

func loadState(ctx context.Context, repos Repositories, tenantID int64, clock Clock) (model.State, error) {
	plan, now, err := currentPlan(ctx, repos, tenantID, clock)
	if err != nil {
		return model.State{}, err
	}
	settings, err := repos.Settings.Get(ctx, tenantID)
	if err != nil {
		return model.State{}, fmt.Errorf("load settings: %w", err)
	}
	active, err := openSession(ctx, repos, tenantID, settings, now)
	if err != nil {
		return model.State{}, err
	}
	weeklySignals, err := repos.Sessions.SumChasingSignalsSince(ctx, tenantID, plan.WeekStart)
	if err != nil {
		return model.State{}, fmt.Errorf("count chasing signals: %w", err)
	}
	return model.State{
		Plan:                   plan,
		ProgressPercent:        plan.ProgressPercent(),
		FrictionLevel:          engine.FrictionLevel(plan, settings, active, weeklySignals),
		ChasingSignalsThisWeek: weeklySignals,
		ActiveSession:          active,
		NextAllocationAt:       engine.NextWeekStart(plan.WeekStart),
	}, nil
}

// currentPlan locks the tenant's plan, serializing its bets, and applies any due week changes.
// The time is read after the lock so a request that waited in a queue sees the real moment it runs.
func currentPlan(ctx context.Context, repos Repositories, tenantID int64, clock Clock) (model.Plan, time.Time, error) {
	plan, err := repos.Plans.GetForUpdate(ctx, tenantID)
	if err != nil {
		return model.Plan{}, time.Time{}, fmt.Errorf("load plan: %w", err)
	}
	now := clock.Now()
	plan.WeekStart = plan.WeekStart.In(now.Location())

	rolled := false
	for !now.Before(engine.NextWeekStart(plan.WeekStart)) {
		chased, err := repos.Sessions.CountEndedBetween(ctx, tenantID, model.ReasonChasingLimit, plan.WeekStart, engine.NextWeekStart(plan.WeekStart))
		if err != nil {
			return model.Plan{}, time.Time{}, fmt.Errorf("evaluate week: %w", err)
		}
		engine.AdvanceWeek(&plan, chased == 0)
		rolled = true
	}
	if !rolled {
		return plan, now, nil
	}
	err = repos.Plans.Save(ctx, plan)
	if err != nil {
		return model.Plan{}, time.Time{}, fmt.Errorf("save plan: %w", err)
	}
	return plan, now, nil
}

// openSession returns the tenant's live session, closing it first if its time ran out.
func openSession(ctx context.Context, repos Repositories, tenantID int64, settings model.RuleSettings, now time.Time) (*model.Session, error) {
	session, err := repos.Sessions.FindOpen(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("find open session: %w", err)
	}
	if session == nil || !engine.SessionExpired(*session, settings, now) {
		return session, nil
	}
	session.End(model.ReasonSessionTimeLimit, now)
	err = repos.Sessions.Save(ctx, *session)
	if err != nil {
		return nil, fmt.Errorf("close expired session: %w", err)
	}
	return nil, nil
}
