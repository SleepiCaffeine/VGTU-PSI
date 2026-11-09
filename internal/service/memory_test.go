package service_test

import (
	"context"
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service"
)

type memory struct {
	plans         map[int64]model.Plan
	settings      map[int64]model.RuleSettings
	sessions      map[int64]model.Session
	confirmations map[int64]model.Confirmation
	bets          []model.Bet
	alerts        []model.Alert
	nextID        int64
}

func newMemory() *memory {
	return &memory{
		plans:         map[int64]model.Plan{},
		settings:      map[int64]model.RuleSettings{},
		sessions:      map[int64]model.Session{},
		confirmations: map[int64]model.Confirmation{},
		nextID:        1000,
	}
}

func (m *memory) id() int64 {
	m.nextID++
	return m.nextID
}

func (m *memory) Do(ctx context.Context, fn func(context.Context, service.Repositories) error) error {
	return fn(ctx, service.Repositories{
		Plans:         memoryPlans{m},
		Settings:      memorySettings{m},
		Sessions:      memorySessions{m},
		Bets:          memoryBets{m},
		Confirmations: memoryConfirmations{m},
		Alerts:        memoryAlerts{m},
	})
}

type memoryPlans struct{ m *memory }

func (r memoryPlans) GetForUpdate(_ context.Context, tenantID int64) (model.Plan, error) {
	plan, found := r.m.plans[tenantID]
	if !found {
		return model.Plan{}, service.ErrNotFound
	}
	return plan, nil
}

func (r memoryPlans) Save(_ context.Context, plan model.Plan) error {
	r.m.plans[plan.TenantID] = plan
	return nil
}

type memorySettings struct{ m *memory }

func (r memorySettings) Get(_ context.Context, tenantID int64) (model.RuleSettings, error) {
	return r.m.settings[tenantID], nil
}

type memorySessions struct{ m *memory }

func (r memorySessions) Get(_ context.Context, tenantID, sessionID int64) (model.Session, error) {
	session, found := r.m.sessions[sessionID]
	if !found || session.TenantID != tenantID {
		return model.Session{}, service.ErrNotFound
	}
	return session, nil
}

func (r memorySessions) FindOpen(_ context.Context, tenantID int64) (*model.Session, error) {
	for _, session := range r.m.sessions {
		if session.TenantID == tenantID && !session.Ended() {
			return &session, nil
		}
	}
	return nil, nil
}

func (r memorySessions) Create(_ context.Context, session *model.Session) error {
	session.ID = r.m.id()
	r.m.sessions[session.ID] = *session
	return nil
}

func (r memorySessions) Save(_ context.Context, session model.Session) error {
	r.m.sessions[session.ID] = session
	return nil
}

func (r memorySessions) CountStartedSince(_ context.Context, tenantID int64, since time.Time) (int, error) {
	count := 0
	for _, session := range r.m.sessions {
		if session.TenantID == tenantID && !session.StartedAt.Before(since) {
			count++
		}
	}
	return count, nil
}

func (r memorySessions) LastEndedAt(context.Context, int64) (*time.Time, error) {
	return nil, nil
}

func (r memorySessions) CountEndedBetween(_ context.Context, tenantID int64, reason model.Reason, from, to time.Time) (int, error) {
	count := 0
	for _, session := range r.m.sessions {
		if session.TenantID == tenantID && session.EndReason == reason && !session.EndedAt.Before(from) && session.EndedAt.Before(to) {
			count++
		}
	}
	return count, nil
}

func (r memorySessions) ShiftBack(context.Context, int64, time.Duration) error {
	return nil
}

type memoryBets struct{ m *memory }

func (r memoryBets) Create(_ context.Context, bet *model.Bet) error {
	r.m.bets = append(r.m.bets, *bet)
	return nil
}

type memoryConfirmations struct{ m *memory }

func (r memoryConfirmations) Find(_ context.Context, tenantID, confirmationID int64) (*model.Confirmation, error) {
	confirmation, found := r.m.confirmations[confirmationID]
	if !found || confirmation.TenantID != tenantID {
		return nil, nil
	}
	return &confirmation, nil
}

func (r memoryConfirmations) Create(_ context.Context, confirmation *model.Confirmation) error {
	confirmation.ID = r.m.id()
	r.m.confirmations[confirmation.ID] = *confirmation
	return nil
}

func (r memoryConfirmations) Save(_ context.Context, confirmation model.Confirmation) error {
	r.m.confirmations[confirmation.ID] = confirmation
	return nil
}

type memoryAlerts struct{ m *memory }

func (r memoryAlerts) Create(_ context.Context, alert *model.Alert) error {
	r.m.alerts = append(r.m.alerts, *alert)
	return nil
}

func (r memoryAlerts) List(_ context.Context, tenantID int64) ([]model.Alert, error) {
	var alerts []model.Alert
	for _, alert := range r.m.alerts {
		if alert.TenantID == tenantID {
			alerts = append(alerts, alert)
		}
	}
	return alerts, nil
}
