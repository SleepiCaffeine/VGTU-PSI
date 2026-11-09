package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/game"
)

const (
	tenantA = 1
	tenantB = 2
)

var (
	weekStart = time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	wednesday = weekStart.Add(50 * time.Hour)
)

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type blackPocket struct{}

func (blackPocket) IntN(int) int { return 2 }

func deps(m *memory, now time.Time) service.Dependencies {
	return service.Dependencies{
		UnitOfWork: m, Clock: fixedClock{now}, Engine: engine.Default(), Games: game.Default(), Random: blackPocket{},
	}
}

func seedTenant(m *memory, tenantID int64, week, balance int) {
	m.settings[tenantID] = model.RuleSettings{
		TenantID: tenantID, MaxStake: 50, StopLoss: 200, SessionMinutes: 60, WeeklySessions: 5, AllowedToHour: 24,
		ChasingBreakSeconds: 30, ConfirmEveryBets: 5, FrictionWaitSeconds: 10, WeeksPerFrictionLevel: 2, BetsPerFrictionLevel: 25,
	}
	m.plans[tenantID] = model.Plan{
		TenantID: tenantID, InitialAllocation: 500, Allocation: 500, Balance: balance, WeekNumber: week, WeekStart: weekStart,
	}
}

func seedSession(m *memory, tenantID, sessionID int64, mutate func(*model.Session)) {
	session := model.Session{ID: sessionID, TenantID: tenantID, StartedAt: wednesday.Add(-time.Minute)}
	mutate(&session)
	m.sessions[sessionID] = session
}

func noChange(*model.Session) {}

func bet(stake int) model.BetRequest {
	return model.BetRequest{Game: "roulette", Selection: "red", Stake: stake}
}

func TestPlaceBetSettlesLostRound(t *testing.T) {
	m := newMemory()
	seedTenant(m, tenantA, 1, 500)
	seedSession(m, tenantA, 10, noChange)

	result, err := service.NewBettingService(deps(m, wednesday)).PlaceBet(context.Background(), tenantA, 10, bet(20))
	if err != nil {
		t.Fatal(err)
	}

	if result.Decision.Outcome != model.OutcomeAccepted || result.Balance != 480 {
		t.Errorf("result = %s balance %d", result.Decision.Outcome, result.Balance)
	}
	if len(m.bets) != 1 || m.bets[0].Stake != 20 || m.bets[0].Payout != 0 {
		t.Errorf("bets = %+v", m.bets)
	}
	if m.sessions[10].NetLoss != 20 || !m.sessions[10].LastLost {
		t.Errorf("session = %+v", m.sessions[10])
	}
}

func TestChasingLimitPersistsEndAndAlert(t *testing.T) {
	m := newMemory()
	seedTenant(m, tenantA, 1, 480)
	seedSession(m, tenantA, 10, func(s *model.Session) { s.ChasingSignals, s.LastStake, s.LastLost = 2, 20, true })

	result, err := service.NewBettingService(deps(m, wednesday)).PlaceBet(context.Background(), tenantA, 10, bet(40))
	if err != nil {
		t.Fatal(err)
	}

	if result.Decision.Verdict.Reason != model.ReasonChasingLimit || result.Balance != 480 {
		t.Errorf("result = %s balance %d", result.Decision.Verdict.Reason, result.Balance)
	}
	if m.sessions[10].EndReason != model.ReasonChasingLimit || m.sessions[10].ChasingSignals != 3 {
		t.Errorf("session = %+v", m.sessions[10])
	}
	if len(m.alerts) != 1 || m.alerts[0].Reason != model.ReasonChasingLimit {
		t.Errorf("alerts = %+v", m.alerts)
	}
	if len(m.bets) != 0 {
		t.Errorf("rejected bet was recorded: %+v", m.bets)
	}
}

func TestChasingSignalAlertsSupervisor(t *testing.T) {
	m := newMemory()
	seedTenant(m, tenantA, 1, 480)
	seedSession(m, tenantA, 10, func(s *model.Session) { s.BetCount, s.LastStake, s.LastLost = 1, 20, true })

	result, err := service.NewBettingService(deps(m, wednesday)).PlaceBet(context.Background(), tenantA, 10, bet(40))
	if err != nil {
		t.Fatal(err)
	}

	if result.Decision.Outcome != model.OutcomeAccepted {
		t.Fatalf("outcome = %s", result.Decision.Outcome)
	}
	if len(m.alerts) != 1 || m.alerts[0].Reason != model.ReasonChasingSignal {
		t.Errorf("alerts = %+v", m.alerts)
	}
}

func TestConfirmationIsSingleUse(t *testing.T) {
	m := newMemory()
	seedTenant(m, tenantA, 9, 500)
	seedSession(m, tenantA, 10, noChange)
	betting := service.NewBettingService(deps(m, wednesday))
	ctx := context.Background()

	asked, err := betting.PlaceBet(ctx, tenantA, 10, bet(20))
	if err != nil {
		t.Fatal(err)
	}
	if asked.Decision.Outcome != model.OutcomeConfirmationRequired || asked.Decision.Confirmation.ID == 0 {
		t.Fatalf("decision = %+v", asked.Decision)
	}
	id := asked.Decision.Confirmation.ID
	phrase := asked.Decision.Confirmation.Phrase
	afterWait := service.NewBettingService(deps(m, wednesday.Add(2*time.Second)))

	attempts := []struct {
		name        string
		text        string
		wantOutcome model.Outcome
		wantReason  model.Reason
		wantBalance int
	}{
		{"wrong sentence leaves the balance alone", "neteisingas", model.OutcomeConfirmationFailed, model.ReasonConfirmationMismatch, 500},
		{"correct sentence places the bet", phrase, model.OutcomeAccepted, model.ReasonRulesPassed, 480},
		{"same confirmation cannot be reused", phrase, model.OutcomeConfirmationFailed, model.ReasonConfirmationInvalid, 480},
	}

	for _, attempt := range attempts {
		t.Run(attempt.name, func(t *testing.T) {
			request := bet(20)
			request.ConfirmationID = &id
			request.ConfirmationText = attempt.text

			result, err := afterWait.PlaceBet(ctx, tenantA, 10, request)
			if err != nil {
				t.Fatal(err)
			}
			if result.Decision.Outcome != attempt.wantOutcome || result.Decision.Verdict.Reason != attempt.wantReason {
				t.Fatalf("decision = %s/%s", result.Decision.Outcome, result.Decision.Verdict.Reason)
			}
			if result.Balance != attempt.wantBalance {
				t.Errorf("balance = %d, want %d", result.Balance, attempt.wantBalance)
			}
		})
	}
}

func TestWeekRollover(t *testing.T) {
	chasedAt := weekStart.Add(time.Hour)

	tests := []struct {
		name           string
		now            time.Time
		chased         bool
		wantAllocation int
		wantWeek       int
		wantProgress   int
	}{
		{"same week changes nothing", wednesday, false, 500, 1, 0},
		{"clean week reduces", weekStart.AddDate(0, 0, 7), false, 420, 2, 16},
		{"chasing week keeps allocation", weekStart.AddDate(0, 0, 7), true, 500, 2, 0},
		{"skipped weeks are applied in turn", weekStart.AddDate(0, 0, 14), false, 350, 3, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMemory()
			seedTenant(m, tenantA, 1, 100)
			if tt.chased {
				seedSession(m, tenantA, 10, func(s *model.Session) {
					s.End(model.ReasonChasingLimit, chasedAt)
				})
			}

			state, err := service.NewPlanService(deps(m, tt.now)).State(context.Background(), tenantA)
			if err != nil {
				t.Fatal(err)
			}

			if state.Plan.Allocation != tt.wantAllocation || state.Plan.WeekNumber != tt.wantWeek || state.ProgressPercent != tt.wantProgress {
				t.Errorf("state = %+v", state)
			}
		})
	}
}

func TestStartSession(t *testing.T) {
	tests := []struct {
		name        string
		balance     int
		wantOutcome model.Outcome
		wantReason  model.Reason
		wantSession bool
	}{
		{"credits available", 100, model.OutcomeAccepted, model.ReasonRulesPassed, true},
		{"no credits", 0, model.OutcomeDenied, model.ReasonNoCredits, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMemory()
			seedTenant(m, tenantA, 1, tt.balance)

			result, err := service.NewSessionService(deps(m, wednesday)).Start(context.Background(), tenantA)
			if err != nil {
				t.Fatal(err)
			}

			if result.Decision.Outcome() != tt.wantOutcome || result.Decision.Verdict.Reason != tt.wantReason {
				t.Errorf("decision = %s/%s", result.Decision.Outcome(), result.Decision.Verdict.Reason)
			}
			if (result.Session != nil) != tt.wantSession {
				t.Errorf("session = %+v", result.Session)
			}
		})
	}
}

func TestRestartCannotDodgeChasingRules(t *testing.T) {
	breakEnds := wednesday.Add(time.Minute)
	nextMonday := weekStart.AddDate(0, 0, 7)

	tests := []struct {
		name       string
		mutate     func(*model.Session)
		wantReason model.Reason
		wantNext   *time.Time
	}{
		{"ending a session does not skip its break", func(s *model.Session) {
			s.BreakUntil = &breakEnds
			s.End(model.ReasonUserEnded, wednesday.Add(-time.Second))
		}, model.ReasonBreakActive, nil},
		{"a chasing stop locks the rest of the week", func(s *model.Session) {
			s.End(model.ReasonChasingLimit, wednesday.Add(-time.Second))
		}, model.ReasonChasingLockout, &nextMonday},
		{"last week's stop no longer applies", func(s *model.Session) {
			s.End(model.ReasonChasingLimit, weekStart.Add(-time.Hour))
		}, model.ReasonRulesPassed, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMemory()
			seedTenant(m, tenantA, 1, 100)
			seedSession(m, tenantA, 10, tt.mutate)

			result, err := service.NewSessionService(deps(m, wednesday)).Start(context.Background(), tenantA)
			if err != nil {
				t.Fatal(err)
			}

			if result.Decision.Verdict.Reason != tt.wantReason {
				t.Fatalf("reason = %s, want %s", result.Decision.Verdict.Reason, tt.wantReason)
			}
			if tt.wantNext != nil && (result.Decision.NextAllocationAt == nil || !result.Decision.NextAllocationAt.Equal(*tt.wantNext)) {
				t.Errorf("next allocation = %v, want %v", result.Decision.NextAllocationAt, tt.wantNext)
			}
		})
	}
}

func TestChasingSignalsCarryAcrossSessions(t *testing.T) {
	m := newMemory()
	seedTenant(m, tenantA, 1, 480)
	seedSession(m, tenantA, 10, func(s *model.Session) {
		s.ChasingSignals = 2
		s.End(model.ReasonUserEnded, wednesday.Add(-time.Second))
	})
	seedSession(m, tenantA, 11, func(s *model.Session) { s.BetCount, s.LastStake, s.LastLost = 1, 20, true })

	result, err := service.NewBettingService(deps(m, wednesday)).PlaceBet(context.Background(), tenantA, 11, bet(40))
	if err != nil {
		t.Fatal(err)
	}

	if result.Decision.Verdict.Reason != model.ReasonChasingLimit {
		t.Errorf("reason = %s, want CHASING_LIMIT", result.Decision.Verdict.Reason)
	}
	if m.sessions[11].EndReason != model.ReasonChasingLimit {
		t.Errorf("session = %+v", m.sessions[11])
	}
}

func TestExpiredSessionIsClosedBeforeStarting(t *testing.T) {
	m := newMemory()
	seedTenant(m, tenantA, 1, 100)
	seedSession(m, tenantA, 10, func(s *model.Session) { s.StartedAt = wednesday.Add(-2 * time.Hour) })

	result, err := service.NewSessionService(deps(m, wednesday)).Start(context.Background(), tenantA)
	if err != nil {
		t.Fatal(err)
	}

	if result.Session == nil {
		t.Fatalf("session was not started: %+v", result.Decision)
	}
	if m.sessions[10].EndReason != model.ReasonSessionTimeLimit {
		t.Errorf("old session = %+v", m.sessions[10])
	}
}

func TestBetInputErrors(t *testing.T) {
	tests := []struct {
		name      string
		tenantID  int64
		sessionID int64
		request   model.BetRequest
		wantErr   error
	}{
		{"unknown game", tenantA, 10, model.BetRequest{Game: "keno", Selection: "red", Stake: 10}, game.ErrUnknownGame},
		{"invalid selection", tenantA, 10, model.BetRequest{Game: "roulette", Selection: "green", Stake: 10}, game.ErrInvalidSelection},
		{"session of another tenant", tenantB, 10, bet(10), service.ErrNotFound},
		{"unknown tenant", 99, 10, bet(10), service.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newMemory()
			seedTenant(m, tenantA, 1, 500)
			seedTenant(m, tenantB, 1, 500)
			seedSession(m, tenantA, 10, noChange)

			_, err := service.NewBettingService(deps(m, wednesday)).PlaceBet(context.Background(), tt.tenantID, tt.sessionID, tt.request)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
			if m.plans[tenantA].Balance != 500 || len(m.bets) != 0 {
				t.Errorf("state changed: %+v, %+v", m.plans[tenantA], m.bets)
			}
		})
	}
}
