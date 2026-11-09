// Package service coordinates the rule engine, games and storage.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/game"
)

// ErrNotFound means the record does not exist within the caller's tenant.
var ErrNotFound = errors.New("not found")

type Clock interface {
	Now() time.Time
}

type TenantRepository interface {
	List(ctx context.Context) ([]model.Tenant, error)
}

type PlanRepository interface {
	GetForUpdate(ctx context.Context, tenantID int64) (model.Plan, error)
	Save(ctx context.Context, plan model.Plan) error
}

type SettingsRepository interface {
	Get(ctx context.Context, tenantID int64) (model.RuleSettings, error)
}

type SessionRepository interface {
	Get(ctx context.Context, tenantID, sessionID int64) (model.Session, error)
	FindOpen(ctx context.Context, tenantID int64) (*model.Session, error)
	Create(ctx context.Context, session *model.Session) error
	Save(ctx context.Context, session model.Session) error
	CountStartedSince(ctx context.Context, tenantID int64, since time.Time) (int, error)
	LastEndedAt(ctx context.Context, tenantID int64) (*time.Time, error)
	CountEndedBetween(ctx context.Context, tenantID int64, reason model.Reason, from, to time.Time) (int, error)
	ShiftBack(ctx context.Context, tenantID int64, by time.Duration) error
	SumChasingSignalsSince(ctx context.Context, tenantID int64, since time.Time) (int, error)
	LatestBreakUntil(ctx context.Context, tenantID int64) (*time.Time, error)
}

type BetRepository interface {
	Create(ctx context.Context, bet *model.Bet) error
}

type ConfirmationRepository interface {
	Find(ctx context.Context, tenantID, confirmationID int64) (*model.Confirmation, error)
	Create(ctx context.Context, confirmation *model.Confirmation) error
	Save(ctx context.Context, confirmation model.Confirmation) error
}

type AlertRepository interface {
	Create(ctx context.Context, alert *model.Alert) error
	List(ctx context.Context, tenantID int64) ([]model.Alert, error)
}

// Repositories groups the stores available inside one unit of work.
type Repositories struct {
	Tenants       TenantRepository
	Plans         PlanRepository
	Settings      SettingsRepository
	Sessions      SessionRepository
	Bets          BetRepository
	Confirmations ConfirmationRepository
	Alerts        AlertRepository
}

// UnitOfWork runs fn atomically; returning an error rolls everything back.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

// Dependencies are shared by every service.
type Dependencies struct {
	UnitOfWork UnitOfWork
	Clock      Clock
	Engine     *engine.Engine
	Games      *game.Registry
	Random     game.RandomSource
}
