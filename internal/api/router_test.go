package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/api"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/game"
)

type fakeBetting struct {
	result model.BetResult
	err    error
	calls  int
}

func (f *fakeBetting) PlaceBet(context.Context, int64, int64, model.BetRequest) (model.BetResult, error) {
	f.calls++
	return f.result, f.err
}

type fakePlans struct{}

func (fakePlans) State(context.Context, int64) (model.State, error) {
	return model.State{Plan: model.Plan{Balance: 480}}, nil
}

func (fakePlans) AdvanceWeek(context.Context, int64) (model.State, error) {
	return model.State{}, nil
}

func newRouter(betting *fakeBetting, demoMode bool) http.Handler {
	return api.NewRouter(
		api.Services{Plans: fakePlans{}, Betting: betting},
		api.Options{DemoMode: demoMode, Static: fstest.MapFS{}},
	)
}

const betBody = `{"game":"roulette","selection":"red","stake":20}`

func TestBetStatusMapping(t *testing.T) {
	tests := []struct {
		name       string
		result     model.BetResult
		err        error
		wantStatus int
	}{
		{"accepted", model.BetResult{Decision: model.BetDecision{Outcome: model.OutcomeAccepted, Verdict: model.Verdict{Allowed: true}}}, nil, http.StatusOK},
		{"capped", model.BetResult{Decision: model.BetDecision{Outcome: model.OutcomeAcceptedCapped, Verdict: model.Verdict{Allowed: true}}}, nil, http.StatusOK},
		{"rejected", model.BetResult{Decision: model.BetDecision{Outcome: model.OutcomeRejected}}, nil, http.StatusConflict},
		{"confirmation required", model.BetResult{Decision: model.BetDecision{Outcome: model.OutcomeConfirmationRequired}}, nil, http.StatusConflict},
		{"confirmation failed", model.BetResult{Decision: model.BetDecision{Outcome: model.OutcomeConfirmationFailed}}, nil, http.StatusConflict},
		{"missing session", model.BetResult{}, fmt.Errorf("load: %w", service.ErrNotFound), http.StatusNotFound},
		{"unknown game", model.BetResult{}, game.ErrUnknownGame, http.StatusBadRequest},
		{"invalid selection", model.BetResult{}, game.ErrInvalidSelection, http.StatusBadRequest},
		{"unexpected failure", model.BetResult{}, errors.New("db down"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newRouter(&fakeBetting{result: tt.result, err: tt.err}, false)
			request := httptest.NewRequest(http.MethodPost, "/api/tenants/1/sessions/2/bets", strings.NewReader(betBody))
			request.Header.Set("X-Tenant-ID", "1")
			request.Header.Set("Content-Type", "application/json")
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d: %s", recorder.Code, tt.wantStatus, recorder.Body)
			}
		})
	}
}

func TestInternalErrorsAreNotExposed(t *testing.T) {
	router := newRouter(&fakeBetting{err: errors.New("db password leaked")}, false)
	request := httptest.NewRequest(http.MethodPost, "/api/tenants/1/sessions/2/bets", strings.NewReader(betBody))
	request.Header.Set("X-Tenant-ID", "1")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if strings.Contains(recorder.Body.String(), "leaked") {
		t.Errorf("body exposes internal error: %s", recorder.Body)
	}
}

func TestTenantContextIsEnforced(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		header     string
		wantStatus int
	}{
		{"own tenant", http.MethodGet, "/api/tenants/1/state", "1", http.StatusOK},
		{"other tenant state", http.MethodGet, "/api/tenants/2/state", "1", http.StatusForbidden},
		{"missing context", http.MethodGet, "/api/tenants/1/state", "", http.StatusForbidden},
		{"other tenant bet", http.MethodPost, "/api/tenants/2/sessions/5/bets", "1", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			betting := &fakeBetting{}
			request := httptest.NewRequest(tt.method, tt.path, strings.NewReader(betBody))
			request.Header.Set("X-Tenant-ID", tt.header)
			recorder := httptest.NewRecorder()

			newRouter(betting, false).ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusForbidden && betting.calls != 0 {
				t.Errorf("service was called %d times for a forbidden request", betting.calls)
			}
		})
	}
}

func TestStateResponseShape(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/tenants/1/state", nil)
	request.Header.Set("X-Tenant-ID", "1")
	recorder := httptest.NewRecorder()

	newRouter(&fakeBetting{}, false).ServeHTTP(recorder, request)

	var body map[string]any
	err := json.Unmarshal(recorder.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body["balance"] != 480.0 {
		t.Errorf("balance = %v, want 480", body["balance"])
	}
}

func TestAdvanceWeekOnlyInDemoMode(t *testing.T) {
	tests := []struct {
		name       string
		demoMode   bool
		wantStatus int
	}{
		{"enabled", true, http.StatusOK},
		{"disabled", false, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/tenants/1/demo/advance-week", nil)
			request.Header.Set("X-Tenant-ID", "1")
			recorder := httptest.NewRecorder()

			newRouter(&fakeBetting{}, tt.demoMode).ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
		})
	}
}
