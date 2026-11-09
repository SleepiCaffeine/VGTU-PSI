package engine_test

import (
	"testing"
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
)

const rouletteEdge = 1.0 / 37

var (
	monday = time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	now    = monday.Add(50 * time.Hour)
)

func settings() model.RuleSettings {
	return model.RuleSettings{
		MaxStake: 50, StopLoss: 200, SessionMinutes: 60, WeeklySessions: 5, CooldownMinutes: 30,
		AllowedFromHour: 0, AllowedToHour: 24, ChasingBreakSeconds: 60, ConfirmEveryBets: 5,
		FrictionWaitSeconds: 10, WeeksPerFrictionLevel: 2, BetsPerFrictionLevel: 25,
	}
}

func plan(balance, week int) model.Plan {
	return model.Plan{InitialAllocation: 500, Allocation: 500, Balance: balance, WeekNumber: week, WeekStart: monday}
}

func openSession() model.Session {
	return model.Session{ID: 1, StartedAt: now.Add(-time.Minute)}
}

func decideBet(p model.Plan, s model.Session, stake int) model.BetDecision {
	return engine.Default().DecideBet(engine.BetInput{
		Now: now, Settings: settings(), Plan: p, Session: s, RequestedStake: stake, HouseEdge: rouletteEdge,
		Week: model.WeekActivity{ChasingSignals: s.ChasingSignals},
	})
}

func attemptOf(submitted bool, record *model.Confirmation, text string) *engine.ConfirmationAttempt {
	if !submitted {
		return nil
	}
	return &engine.ConfirmationAttempt{Record: record, Text: text}
}

func TestDesignScenariosBets(t *testing.T) {
	tests := []struct {
		name          string
		plan          model.Plan
		mutate        func(*model.Session)
		stake         int
		wantOutcome   model.Outcome
		wantReason    model.Reason
		wantStake     int
		wantBalance   int
		wantEndReason model.Reason
		wantNotify    bool
	}{
		{
			name:        "usual bet is accepted without confirmation",
			plan:        plan(500, 1),
			mutate:      func(*model.Session) {},
			stake:       20,
			wantOutcome: model.OutcomeAccepted,
			wantReason:  model.ReasonRulesPassed,
			wantStake:   20,
			wantBalance: 480,
		},
		{
			name:          "stake is capped at the remaining stop-loss and ends the session",
			plan:          plan(310, 1),
			mutate:        func(s *model.Session) { s.NetLoss = 190 },
			stake:         20,
			wantOutcome:   model.OutcomeAcceptedCapped,
			wantReason:    model.ReasonStopLossCap,
			wantStake:     10,
			wantBalance:   300,
			wantEndReason: model.ReasonStopLossReached,
		},
		{
			name: "third chasing signal rejects the bet and alerts the supervisor",
			plan: plan(480, 1),
			mutate: func(s *model.Session) {
				s.ChasingSignals, s.LastStake, s.LastLost = 2, 20, true
			},
			stake:         40,
			wantOutcome:   model.OutcomeRejected,
			wantReason:    model.ReasonChasingLimit,
			wantBalance:   480,
			wantEndReason: model.ReasonChasingLimit,
			wantNotify:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.plan
			s := openSession()
			tt.mutate(&s)

			d := decideBet(p, s, tt.stake)

			if d.Outcome != tt.wantOutcome || d.Verdict.Reason != tt.wantReason || d.Stake != tt.wantStake {
				t.Fatalf("decision = %s/%s/%d, want %s/%s/%d", d.Outcome, d.Verdict.Reason, d.Stake, tt.wantOutcome, tt.wantReason, tt.wantStake)
			}
			if d.Confirmation != nil {
				t.Fatalf("unexpected confirmation request: %+v", d.Confirmation)
			}
			if (d.Verdict.Effects.AlertReason != "") != tt.wantNotify {
				t.Errorf("alert reason = %q, want alert %v", d.Verdict.Effects.AlertReason, tt.wantNotify)
			}

			if d.Accepted() {
				const lostRoundPayout = 0
				p, s = engine.SettleRound(p, s, settings(), d.Stake, lostRoundPayout, now)
			}
			s = engine.ApplyEffects(s, d.Verdict.Effects, now)

			if p.Balance != tt.wantBalance {
				t.Errorf("balance = %d, want %d", p.Balance, tt.wantBalance)
			}
			if s.EndReason != tt.wantEndReason {
				t.Errorf("end reason = %q, want %q", s.EndReason, tt.wantEndReason)
			}
		})
	}
}

func TestStakeCaps(t *testing.T) {
	tests := []struct {
		name        string
		balance     int
		netLoss     int
		stake       int
		wantOutcome model.Outcome
		wantReason  model.Reason
		wantStake   int
	}{
		{"max stake caps", 500, 0, 80, model.OutcomeAcceptedCapped, model.ReasonMaxStakeCap, 50},
		{"balance caps", 30, 0, 40, model.OutcomeAcceptedCapped, model.ReasonBalanceCap, 30},
		{"stake at the limit is untouched", 500, 0, 50, model.OutcomeAccepted, model.ReasonRulesPassed, 50},
		{"nothing left rejects", 500, 200, 10, model.OutcomeRejected, model.ReasonNoStakeAvailable, 0},
		{"non-positive stake rejects", 500, 0, 0, model.OutcomeRejected, model.ReasonInvalidStake, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := openSession()
			s.NetLoss = tt.netLoss
			d := decideBet(plan(tt.balance, 1), s, tt.stake)
			if d.Outcome != tt.wantOutcome || d.Verdict.Reason != tt.wantReason || d.Stake != tt.wantStake {
				t.Errorf("decision = %s/%s/%d, want %s/%s/%d", d.Outcome, d.Verdict.Reason, d.Stake, tt.wantOutcome, tt.wantReason, tt.wantStake)
			}
		})
	}
}

func TestConfirmationVerification(t *testing.T) {
	phrase := engine.LossPhrase(20, rouletteEdge)
	used := now.Add(-time.Second)
	valid := model.Confirmation{SessionID: 1, Stake: 20, Kind: model.ConfirmationText, Phrase: phrase, NotBefore: now.Add(-time.Second)}

	alreadyUsed := valid
	alreadyUsed.UsedAt = &used
	tooEarly := valid
	tooEarly.NotBefore = now.Add(time.Minute)
	otherStake := valid
	otherStake.Stake = 30

	tests := []struct {
		name         string
		submitted    bool
		confirmation *model.Confirmation
		text         string
		wantOutcome  model.Outcome
		wantReason   model.Reason
	}{
		{"nothing submitted asks for confirmation", false, nil, "", model.OutcomeConfirmationRequired, model.ReasonConfirmationNeeded},
		{"wrong sentence fails", true, &valid, "neteisingas sakinys", model.OutcomeConfirmationFailed, model.ReasonConfirmationMismatch},
		{"unknown confirmation fails", true, nil, phrase, model.OutcomeConfirmationFailed, model.ReasonConfirmationInvalid},
		{"used confirmation fails", true, &alreadyUsed, phrase, model.OutcomeConfirmationFailed, model.ReasonConfirmationInvalid},
		{"confirmation for another stake fails", true, &otherStake, phrase, model.OutcomeConfirmationFailed, model.ReasonConfirmationInvalid},
		{"confirmation before the wait fails", true, &tooEarly, phrase, model.OutcomeConfirmationFailed, model.ReasonConfirmationTooEarly},
		{"correct sentence is accepted", true, &valid, phrase, model.OutcomeAccepted, model.ReasonRulesPassed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := openSession()
			d := engine.Default().DecideBet(engine.BetInput{
				Now: now, Settings: settings(), Plan: plan(500, 9), Session: s, RequestedStake: 20, HouseEdge: rouletteEdge,
				Attempt: attemptOf(tt.submitted, tt.confirmation, tt.text),
			})
			if d.Outcome != tt.wantOutcome || d.Verdict.Reason != tt.wantReason {
				t.Fatalf("decision = %s/%s, want %s/%s", d.Outcome, d.Verdict.Reason, tt.wantOutcome, tt.wantReason)
			}
		})
	}
}

func TestSessionStartDecisions(t *testing.T) {
	recentEnd := now.Add(-time.Minute)
	open := openSession()
	nextMonday := monday.AddDate(0, 0, 7)
	breakEnds := now.Add(time.Minute)

	tests := []struct {
		name        string
		mutate      func(*engine.SessionStartInput)
		wantOutcome model.Outcome
		wantReason  model.Reason
		wantNext    *time.Time
	}{
		{"allowed", func(*engine.SessionStartInput) {}, model.OutcomeAccepted, model.ReasonRulesPassed, nil},
		{"no credits shows next allocation", func(in *engine.SessionStartInput) { in.Plan.Balance = 0 }, model.OutcomeDenied, model.ReasonNoCredits, &nextMonday},
		{"open session", func(in *engine.SessionStartInput) { in.OpenSession = &open }, model.OutcomeDenied, model.ReasonSessionAlreadyActive, nil},
		{"outside hours", func(in *engine.SessionStartInput) { in.Settings.AllowedToHour = 1 }, model.OutcomeDenied, model.ReasonOutsideAllowedHours, nil},
		{"weekly limit", func(in *engine.SessionStartInput) { in.Week.SessionsStarted = 5 }, model.OutcomeDenied, model.ReasonWeeklySessionLimit, nil},
		{"cooldown", func(in *engine.SessionStartInput) { in.Week.LastSessionEndedAt = &recentEnd }, model.OutcomeDenied, model.ReasonCooldownActive, nil},
		{"chasing stop locks the week", func(in *engine.SessionStartInput) { in.Week.ChasingStops = 1 }, model.OutcomeDenied, model.ReasonChasingLockout, &nextMonday},
		{"running break blocks a restart", func(in *engine.SessionStartInput) { in.Week.LatestBreakUntil = &breakEnds }, model.OutcomeDenied, model.ReasonBreakActive, nil},
		{"finished break does not block", func(in *engine.SessionStartInput) { in.Week.LatestBreakUntil = &recentEnd }, model.OutcomeAccepted, model.ReasonRulesPassed, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := engine.SessionStartInput{Now: now, Settings: settings(), Plan: plan(100, 1)}
			tt.mutate(&in)
			d := engine.Default().DecideSessionStart(in)
			if d.Outcome() != tt.wantOutcome || d.Verdict.Reason != tt.wantReason {
				t.Fatalf("decision = %s/%s, want %s/%s", d.Outcome(), d.Verdict.Reason, tt.wantOutcome, tt.wantReason)
			}
			if tt.wantNext != nil && (d.NextAllocationAt == nil || !d.NextAllocationAt.Equal(*tt.wantNext)) {
				t.Errorf("next allocation = %v, want %v", d.NextAllocationAt, tt.wantNext)
			}
		})
	}
}

func TestPipelineOrder(t *testing.T) {
	breakUntil := now.Add(time.Minute)
	ended := now.Add(-time.Minute)

	tests := []struct {
		name       string
		mutate     func(*model.Session)
		stake      int
		wantReason model.Reason
	}{
		{"ended session wins over chasing", func(s *model.Session) {
			s.EndedAt, s.ChasingSignals, s.LastStake, s.LastLost = &ended, 2, 20, true
		}, 40, model.ReasonSessionEnded},
		{"time limit wins over break", func(s *model.Session) {
			s.StartedAt, s.BreakUntil = now.Add(-2*time.Hour), &breakUntil
		}, 20, model.ReasonSessionTimeLimit},
		{"break wins over chasing", func(s *model.Session) {
			s.BreakUntil, s.ChasingSignals, s.LastStake, s.LastLost = &breakUntil, 2, 20, true
		}, 40, model.ReasonChasingBreak},
		{"chasing wins over stake limits", func(s *model.Session) {
			s.ChasingSignals, s.LastStake, s.LastLost, s.NetLoss = 2, 20, true, 200
		}, 40, model.ReasonChasingLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := openSession()
			tt.mutate(&s)
			d := decideBet(plan(500, 1), s, tt.stake)
			if d.Outcome != model.OutcomeRejected || d.Verdict.Reason != tt.wantReason {
				t.Fatalf("decision = %s/%s, want REJECTED/%s", d.Outcome, d.Verdict.Reason, tt.wantReason)
			}
		})
	}
}

func TestChasingSignalsBelowLimit(t *testing.T) {
	tests := []struct {
		name       string
		signals    int
		stake      int
		wantSignal bool
		wantBreak  bool
	}{
		{"first signal allows the bet and starts a break", 0, 40, true, true},
		{"second signal allows the bet and starts a break", 1, 40, true, true},
		{"just under double is not chasing", 0, 39, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := openSession()
			s.BetCount, s.ChasingSignals, s.LastStake, s.LastLost = 1, tt.signals, 20, true
			d := decideBet(plan(500, 1), s, tt.stake)
			if d.Outcome != model.OutcomeAccepted {
				t.Fatalf("outcome = %s, want ACCEPTED", d.Outcome)
			}
			if d.Verdict.Effects.ChasingSignal != tt.wantSignal || (d.Verdict.Effects.BreakUntil != nil) != tt.wantBreak ||
				(d.Verdict.Effects.AlertReason == model.ReasonChasingSignal) != tt.wantSignal {
				t.Errorf("effects = %+v, want signal %v break %v", d.Verdict.Effects, tt.wantSignal, tt.wantBreak)
			}
		})
	}
}

func TestChasingLimitCountsTheWholeWeek(t *testing.T) {
	s := openSession()
	s.BetCount, s.LastStake, s.LastLost = 1, 20, true

	d := engine.Default().DecideBet(engine.BetInput{
		Now: now, Settings: settings(), Plan: plan(500, 1), Session: s, RequestedStake: 40, HouseEdge: rouletteEdge,
		Week: model.WeekActivity{ChasingSignals: 2},
	})

	if d.Outcome != model.OutcomeRejected || d.Verdict.Reason != model.ReasonChasingLimit {
		t.Fatalf("decision = %s/%s, want REJECTED/CHASING_LIMIT", d.Outcome, d.Verdict.Reason)
	}
}

func TestConfirmationStopDropsChasingEffects(t *testing.T) {
	s := openSession()
	s.LastStake, s.LastLost = 20, true

	d := decideBet(plan(500, 9), s, 40)

	if d.Outcome != model.OutcomeConfirmationRequired {
		t.Fatalf("outcome = %s, want CONFIRMATION_REQUIRED", d.Outcome)
	}
	if d.Verdict.Effects != (model.Effects{}) {
		t.Errorf("effects kept on confirmation stop: %+v", d.Verdict.Effects)
	}
}

func TestFrictionRequirements(t *testing.T) {
	tests := []struct {
		name     string
		week     int
		betCount int
		wantKind model.ConfirmationKind
		wantWait int
	}{
		{"level 0 needs nothing", 1, 0, "", 0},
		{"level 1 clicks at session start", 3, 0, model.ConfirmationClick, 1},
		{"level 1 needs nothing later", 3, 1, "", 0},
		{"level 2 clicks regularly", 5, 5, model.ConfirmationClick, 1},
		{"level 2 skips between intervals", 5, 6, "", 0},
		{"level 3 types and waits regularly", 7, 5, model.ConfirmationText, 10},
		{"level 3 skips between intervals", 7, 6, "", 0},
		{"level 4 types on every bet", 9, 6, model.ConfirmationText, 1},
		{"level is capped at 4", 99, 7, model.ConfirmationText, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := openSession()
			s.BetCount = tt.betCount
			d := decideBet(plan(500, tt.week), s, 20)
			if tt.wantKind == "" {
				if d.Confirmation != nil {
					t.Fatalf("unexpected confirmation %+v", d.Confirmation)
				}
				return
			}
			if d.Confirmation == nil {
				t.Fatalf("no confirmation requested, outcome %s", d.Outcome)
			}
			if d.Confirmation.Kind != tt.wantKind || d.Confirmation.WaitSeconds != tt.wantWait {
				t.Errorf("confirmation = %+v, want kind %s wait %d", d.Confirmation, tt.wantKind, tt.wantWait)
			}
		})
	}
}

func TestLossPhraseMatchesDesignExample(t *testing.T) {
	want := "Šis statymas yra nuostolingas: vidutiniškai prarasiu 0,54 taško."
	got := engine.LossPhrase(20, rouletteEdge)
	if got != want {
		t.Errorf("phrase = %q, want %q", got, want)
	}
}

func TestPlanReduction(t *testing.T) {
	tests := []struct {
		name           string
		allocation     int
		clean          bool
		wantAllocation int
		wantProgress   int
	}{
		{"clean week reduces 500 to 420", 500, true, 420, 16},
		{"unclean week keeps allocation", 500, false, 500, 0},
		{"rounds down to tens", 420, true, 350, 30},
		{"small allocation reaches zero", 10, true, 0, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := plan(0, 1)
			p.Allocation = tt.allocation
			p = engine.AdvanceWeek(p, tt.clean)
			if p.Allocation != tt.wantAllocation || p.Balance != tt.wantAllocation {
				t.Errorf("allocation/balance = %d/%d, want %d", p.Allocation, p.Balance, tt.wantAllocation)
			}
			if p.ProgressPercent() != tt.wantProgress {
				t.Errorf("progress = %d, want %d", p.ProgressPercent(), tt.wantProgress)
			}
			if p.WeekNumber != 2 || !p.WeekStart.Equal(monday.AddDate(0, 0, 7)) {
				t.Errorf("week = %d starting %v", p.WeekNumber, p.WeekStart)
			}
		})
	}
}

func TestWeekStartOf(t *testing.T) {
	tests := []struct {
		name string
		at   time.Time
	}{
		{"monday midnight", monday},
		{"wednesday", now},
		{"sunday night", monday.Add(6*24*time.Hour + 23*time.Hour)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.WeekStartOf(tt.at)
			if !got.Equal(monday) {
				t.Errorf("week start = %v, want %v", got, monday)
			}
		})
	}
}
