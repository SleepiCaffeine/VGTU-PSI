package service

import (
	"context"
	"fmt"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
)

type SessionService struct {
	deps Dependencies
}

func NewSessionService(deps Dependencies) *SessionService {
	return &SessionService{deps: deps}
}

func (s *SessionService) Start(ctx context.Context, tenantID int64) (model.SessionStartResult, error) {
	var result model.SessionStartResult
	err := s.deps.UnitOfWork.Do(ctx, func(ctx context.Context, repos Repositories) error {
		var err error
		result, err = s.start(ctx, repos, tenantID)
		return err
	})
	return result, err
}

func (s *SessionService) start(ctx context.Context, repos Repositories, tenantID int64) (model.SessionStartResult, error) {
	plan, now, err := currentPlan(ctx, repos, tenantID, s.deps.Clock)
	if err != nil {
		return model.SessionStartResult{}, err
	}
	settings, err := repos.Settings.Get(ctx, tenantID)
	if err != nil {
		return model.SessionStartResult{}, fmt.Errorf("load settings: %w", err)
	}
	open, err := openSession(ctx, repos, tenantID, settings, now)
	if err != nil {
		return model.SessionStartResult{}, err
	}
	week, err := repos.History.WeekActivity(ctx, tenantID, plan.WeekStart, engine.NextWeekStart(plan.WeekStart))
	if err != nil {
		return model.SessionStartResult{}, fmt.Errorf("load week activity: %w", err)
	}

	decision := s.deps.Engine.DecideSessionStart(engine.SessionStartInput{
		Now: now, Settings: settings, Plan: plan, OpenSession: open, Week: week,
	})
	if !decision.Accepted() {
		return model.SessionStartResult{Decision: decision}, nil
	}

	session := model.Session{TenantID: tenantID, StartedAt: now}
	err = repos.Sessions.Create(ctx, &session)
	if err != nil {
		return model.SessionStartResult{}, fmt.Errorf("create session: %w", err)
	}
	return model.SessionStartResult{Decision: decision, Session: &session}, nil
}

func (s *SessionService) End(ctx context.Context, tenantID, sessionID int64) (model.Session, error) {
	var session model.Session
	err := s.deps.UnitOfWork.Do(ctx, func(ctx context.Context, repos Repositories) error {
		var err error
		session, err = repos.Sessions.Get(ctx, tenantID, sessionID)
		if err != nil {
			return fmt.Errorf("load session: %w", err)
		}
		session.End(model.ReasonUserEnded, s.deps.Clock.Now())
		return repos.Sessions.Save(ctx, session)
	})
	return session, err
}
