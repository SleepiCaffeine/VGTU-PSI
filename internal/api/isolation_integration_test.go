//go:build integration

package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"gorm.io/gorm"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/api"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/infra/clock"
	infra "github.com/SleepiCaffeine/VGTU-PSI/internal/infra/mysql"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/game"
)

const foreignBalance = 777

type losingRandom struct{}

// IntN always lands on pocket 2, which is black.
func (losingRandom) IntN(int) int { return 2 }

type pair struct {
	tenantID  int64
	sessionID int64
}

func createPair(t *testing.T, db *gorm.DB, balance int, now time.Time) pair {
	t.Helper()
	tenant := model.Tenant{Name: "integration " + t.Name()}
	must(t, db.Create(&tenant).Error)
	t.Cleanup(func() { deleteTenant(t, db, tenant.ID) })

	settings := model.RuleSettings{
		TenantID: tenant.ID, MaxStake: 50, StopLoss: 200, SessionMinutes: 60, WeeklySessions: 5,
		AllowedToHour: 24, ChasingBreakSeconds: 30, ConfirmEveryBets: 5, FrictionWaitSeconds: 10,
		WeeksPerFrictionLevel: 2, BetsPerFrictionLevel: 25,
	}
	plan := model.Plan{
		TenantID: tenant.ID, InitialAllocation: 500, Allocation: 500, Balance: balance,
		WeekNumber: 1, WeekStart: engine.WeekStartOf(now),
	}
	session := model.Session{TenantID: tenant.ID, StartedAt: now.Add(-time.Minute)}
	must(t, db.Create(&settings).Error)
	must(t, db.Create(&plan).Error)
	must(t, db.Create(&session).Error)
	must(t, db.Create(&model.Alert{TenantID: tenant.ID, SessionID: session.ID, Reason: model.ReasonChasingLimit, CreatedAt: now}).Error)
	return pair{tenantID: tenant.ID, sessionID: session.ID}
}

func deleteTenant(t *testing.T, db *gorm.DB, tenantID int64) {
	t.Helper()
	for _, table := range []any{
		&model.Alert{}, &model.Confirmation{}, &model.Bet{}, &model.Session{}, &model.Plan{}, &model.RuleSettings{}, &model.User{},
	} {
		must(t, db.Where("tenant_id = ?", tenantID).Delete(table).Error)
	}
	must(t, db.Delete(&model.Tenant{}, tenantID).Error)
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

type fixture struct {
	db     *gorm.DB
	router http.Handler
	now    time.Time
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	db, err := infra.Open(infra.DefaultDSN)
	must(t, err)
	must(t, infra.Migrate(db))
	wallClock, err := clock.New()
	must(t, err)

	deps := service.Dependencies{
		UnitOfWork: infra.NewUnitOfWork(db),
		Clock:      wallClock,
		Engine:     engine.Default(),
		Games:      game.Default(),
		Random:     losingRandom{},
	}
	router := api.NewRouter(api.Services{
		Tenants:  service.NewTenantService(deps),
		Plans:    service.NewPlanService(deps),
		Sessions: service.NewSessionService(deps),
		Betting:  service.NewBettingService(deps),
		Alerts:   service.NewAlertService(deps),
	}, api.Options{DemoMode: true, Static: fstest.MapFS{}})
	return fixture{db: db, router: router, now: wallClock.Now()}
}

func (f fixture) call(method, path string, tenantContext int64, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("X-Tenant-ID", fmt.Sprint(tenantContext))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, request)
	return recorder
}

type snapshot struct {
	plan     model.Plan
	session  model.Session
	bets     int64
	sessions int64
	alerts   int64
}

func (f fixture) snapshot(t *testing.T, p pair) snapshot {
	t.Helper()
	var s snapshot
	must(t, f.db.Where("tenant_id = ?", p.tenantID).Take(&s.plan).Error)
	must(t, f.db.Where("tenant_id = ? AND id = ?", p.tenantID, p.sessionID).Take(&s.session).Error)
	must(t, f.db.Model(&model.Bet{}).Where("tenant_id = ?", p.tenantID).Count(&s.bets).Error)
	must(t, f.db.Model(&model.Session{}).Where("tenant_id = ?", p.tenantID).Count(&s.sessions).Error)
	must(t, f.db.Model(&model.Alert{}).Where("tenant_id = ?", p.tenantID).Count(&s.alerts).Error)
	return s
}

func assertUnchanged(t *testing.T, before, after snapshot) {
	t.Helper()
	if before.plan.Balance != after.plan.Balance {
		t.Errorf("balance changed: %d to %d", before.plan.Balance, after.plan.Balance)
	}
	if before.session.BetCount != after.session.BetCount || before.session.Ended() != after.session.Ended() {
		t.Errorf("session changed: %+v to %+v", before.session, after.session)
	}
	if before.bets != after.bets || before.sessions != after.sessions || before.alerts != after.alerts {
		t.Errorf("row counts changed: %+v to %+v", before, after)
	}
}

func TestTenantIsolation(t *testing.T) {
	f := newFixture(t)
	mine := createPair(t, f.db, 500, f.now)
	theirs := createPair(t, f.db, foreignBalance, f.now)
	bet := `{"game":"roulette","selection":"red","stake":20}`

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{"read their state", http.MethodGet, fmt.Sprintf("/api/tenants/%d/state", theirs.tenantID), "", http.StatusForbidden},
		{"read their alerts", http.MethodGet, fmt.Sprintf("/api/tenants/%d/alerts", theirs.tenantID), "", http.StatusForbidden},
		{"start their session", http.MethodPost, fmt.Sprintf("/api/tenants/%d/sessions", theirs.tenantID), "", http.StatusForbidden},
		{"bet on their session", http.MethodPost, fmt.Sprintf("/api/tenants/%d/sessions/%d/bets", theirs.tenantID, theirs.sessionID), bet, http.StatusForbidden},
		{"end their session", http.MethodPost, fmt.Sprintf("/api/tenants/%d/sessions/%d/end", theirs.tenantID, theirs.sessionID), "", http.StatusForbidden},
		{"advance their week", http.MethodPost, fmt.Sprintf("/api/tenants/%d/demo/advance-week", theirs.tenantID), "", http.StatusForbidden},
		{"bet on their session id through my path", http.MethodPost, fmt.Sprintf("/api/tenants/%d/sessions/%d/bets", mine.tenantID, theirs.sessionID), bet, http.StatusNotFound},
		{"end their session id through my path", http.MethodPost, fmt.Sprintf("/api/tenants/%d/sessions/%d/end", mine.tenantID, theirs.sessionID), "", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := f.snapshot(t, theirs)

			response := f.call(tt.method, tt.path, mine.tenantID, tt.body)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d: %s", response.Code, tt.wantStatus, response.Body)
			}
			if strings.Contains(response.Body.String(), fmt.Sprint(foreignBalance)) {
				t.Errorf("response leaks foreign data: %s", response.Body)
			}
			assertUnchanged(t, before, f.snapshot(t, theirs))
		})
	}
}

func TestAlertsAreScopedToOwnTenant(t *testing.T) {
	f := newFixture(t)
	mine := createPair(t, f.db, 500, f.now)
	createPair(t, f.db, foreignBalance, f.now)

	response := f.call(http.MethodGet, fmt.Sprintf("/api/tenants/%d/alerts", mine.tenantID), mine.tenantID, "")

	var alerts []struct {
		SessionID int64 `json:"sessionId"`
	}
	must(t, json.Unmarshal(response.Body.Bytes(), &alerts))
	if len(alerts) != 1 || alerts[0].SessionID != mine.sessionID {
		t.Errorf("alerts = %+v, want only session %d", alerts, mine.sessionID)
	}
}

func TestBetFlowPersists(t *testing.T) {
	f := newFixture(t)
	mine := createPair(t, f.db, 500, f.now)

	response := f.call(http.MethodPost, fmt.Sprintf("/api/tenants/%d/sessions/%d/bets", mine.tenantID, mine.sessionID), mine.tenantID,
		`{"game":"roulette","selection":"red","stake":20}`)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body)
	}
	after := f.snapshot(t, mine)
	if after.plan.Balance != 480 || after.session.NetLoss != 20 || after.session.BetCount != 1 || after.bets != 1 {
		t.Errorf("unexpected state: %+v", after)
	}
}

func TestConcurrentBetsCannotOverspend(t *testing.T) {
	f := newFixture(t)
	mine := createPair(t, f.db, 30, f.now)
	path := fmt.Sprintf("/api/tenants/%d/sessions/%d/bets", mine.tenantID, mine.sessionID)
	const requests = 5
	done := make(chan struct{}, requests)

	for range requests {
		go func() {
			f.call(http.MethodPost, path, mine.tenantID, `{"game":"roulette","selection":"red","stake":10}`)
			done <- struct{}{}
		}()
	}
	for range requests {
		<-done
	}

	var plan model.Plan
	must(t, f.db.WithContext(context.Background()).Where("tenant_id = ?", mine.tenantID).Take(&plan).Error)
	if plan.Balance != 0 {
		t.Errorf("balance = %d, want 0", plan.Balance)
	}
}
