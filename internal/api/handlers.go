package api

import (
	"context"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/api/gen"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service"
)

type handlers struct {
	services Services
	demoMode bool
}

func (h handlers) ListTenants(ctx context.Context, _ gen.ListTenantsRequestObject) (gen.ListTenantsResponseObject, error) {
	tenants, err := h.services.Tenants.List(ctx)
	if err != nil {
		return nil, err
	}
	response := make(gen.ListTenants200JSONResponse, 0, len(tenants))
	for _, tenant := range tenants {
		response = append(response, gen.Tenant{Id: tenant.ID, Name: tenant.Name})
	}
	return response, nil
}

func (h handlers) GetState(ctx context.Context, request gen.GetStateRequestObject) (gen.GetStateResponseObject, error) {
	state, err := h.services.Plans.State(ctx, request.TenantId)
	if err != nil {
		return nil, err
	}
	return gen.GetState200JSONResponse(toState(state)), nil
}

func (h handlers) AdvanceWeek(ctx context.Context, request gen.AdvanceWeekRequestObject) (gen.AdvanceWeekResponseObject, error) {
	if !h.demoMode {
		return nil, service.ErrNotFound
	}
	state, err := h.services.Plans.AdvanceWeek(ctx, request.TenantId)
	if err != nil {
		return nil, err
	}
	return gen.AdvanceWeek200JSONResponse(toState(state)), nil
}

func (h handlers) StartSession(ctx context.Context, request gen.StartSessionRequestObject) (gen.StartSessionResponseObject, error) {
	result, err := h.services.Sessions.Start(ctx, request.TenantId)
	if err != nil {
		return nil, err
	}
	response := toSessionStartResult(result)
	if result.Decision.Accepted() {
		return gen.StartSession201JSONResponse(response), nil
	}
	return gen.StartSession409JSONResponse(response), nil
}

func (h handlers) EndSession(ctx context.Context, request gen.EndSessionRequestObject) (gen.EndSessionResponseObject, error) {
	session, err := h.services.Sessions.End(ctx, request.TenantId, request.SessionId)
	if err != nil {
		return nil, err
	}
	return gen.EndSession200JSONResponse(toSession(session)), nil
}

func (h handlers) PlaceBet(ctx context.Context, request gen.PlaceBetRequestObject) (gen.PlaceBetResponseObject, error) {
	result, err := h.services.Betting.PlaceBet(ctx, request.TenantId, request.SessionId, toBetRequest(*request.Body))
	if err != nil {
		return nil, err
	}
	response := toBetResult(result)
	if result.Decision.Accepted() {
		return gen.PlaceBet200JSONResponse(response), nil
	}
	return gen.PlaceBet409JSONResponse(response), nil
}

func (h handlers) ListAlerts(ctx context.Context, request gen.ListAlertsRequestObject) (gen.ListAlertsResponseObject, error) {
	alerts, err := h.services.Alerts.List(ctx, request.TenantId)
	if err != nil {
		return nil, err
	}
	response := make(gen.ListAlerts200JSONResponse, 0, len(alerts))
	for _, alert := range alerts {
		response = append(response, gen.Alert{
			Id: alert.ID, SessionId: alert.SessionID, Reason: string(alert.Reason), CreatedAt: alert.CreatedAt,
		})
	}
	return response, nil
}

func toBetRequest(body gen.BetRequest) model.BetRequest {
	request := model.BetRequest{
		Game:           body.Game,
		Selection:      body.Selection,
		Stake:          body.Stake,
		ConfirmationID: body.ConfirmationId,
	}
	if body.ConfirmationText != nil {
		request.ConfirmationText = *body.ConfirmationText
	}
	return request
}
