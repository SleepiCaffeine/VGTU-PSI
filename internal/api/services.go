package api

import (
	"context"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type TenantService interface {
	List(ctx context.Context) ([]model.Tenant, error)
}

type PlanService interface {
	State(ctx context.Context, tenantID int64) (model.State, error)
	AdvanceWeek(ctx context.Context, tenantID int64) (model.State, error)
}

type SessionService interface {
	Start(ctx context.Context, tenantID int64) (model.SessionStartResult, error)
	End(ctx context.Context, tenantID, sessionID int64) (model.Session, error)
}

type BettingService interface {
	PlaceBet(ctx context.Context, tenantID, sessionID int64, request model.BetRequest) (model.BetResult, error)
}

type AlertService interface {
	List(ctx context.Context, tenantID int64) ([]model.Alert, error)
}

// Services are the use cases the HTTP layer depends on.
type Services struct {
	Tenants  TenantService
	Plans    PlanService
	Sessions SessionService
	Betting  BettingService
	Alerts   AlertService
}
