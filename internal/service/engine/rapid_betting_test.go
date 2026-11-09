package engine_test

import (
	"testing"
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
)

func rapidSettings() model.RuleSettings {
	s := settings()
	s.RapidBetSeconds, s.RapidBetLimit = 3, 5
	return s
}

func TestRapidBetting(t *testing.T) {
	tests := []struct {
		name        string
		settings    model.RuleSettings
		streak      int
		gap         time.Duration
		wantOutcome model.Outcome
		wantReason  model.Reason
		wantFlag    bool
	}{
		{"streak below the limit is allowed", rapidSettings(), 3, time.Second, model.OutcomeAccepted, model.ReasonRulesPassed, false},
		{"bet that reaches the limit is rejected", rapidSettings(), 4, time.Second, model.OutcomeRejected, model.ReasonRapidBetting, true},
		{"slow bet is allowed despite a long streak", rapidSettings(), 4, 3 * time.Second, model.OutcomeAccepted, model.ReasonRulesPassed, false},
		{"rule is off without a limit", settings(), 40, time.Millisecond, model.OutcomeAccepted, model.ReasonRulesPassed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := openSession()
			last := now.Add(-tt.gap)
			s.BetCount, s.RapidStreak, s.LastBetAt = 1, tt.streak, &last

			d := engine.Default().DecideBet(engine.Input{
				Now: now, Settings: tt.settings, Plan: plan(500, 1), Session: &s, RequestedStake: 20, HouseEdge: rouletteEdge,
			})

			if d.Outcome != tt.wantOutcome || d.Reason != tt.wantReason {
				t.Fatalf("decision = %s/%s, want %s/%s", d.Outcome, d.Reason, tt.wantOutcome, tt.wantReason)
			}
			if d.Effects.RapidFlag != tt.wantFlag || (d.Effects.BreakUntil != nil) != tt.wantFlag {
				t.Errorf("effects = %+v, want flag and break %v", d.Effects, tt.wantFlag)
			}
			if tt.wantFlag && d.Effects.AlertReason != model.ReasonRapidBetting {
				t.Errorf("alert reason = %q", d.Effects.AlertReason)
			}
		})
	}
}

func TestSpammingEqualStakesIsStopped(t *testing.T) {
	p := plan(500, 1)
	s := openSession()
	at := now
	rejectedOnBet := 0

	for attempt := 1; attempt <= 10 && rejectedOnBet == 0; attempt++ {
		d := engine.Default().DecideBet(engine.Input{
			Now: at, Settings: rapidSettings(), Plan: p, Session: &s, RequestedStake: 20, HouseEdge: rouletteEdge,
		})
		if d.Reason == model.ReasonRapidBetting {
			rejectedOnBet = attempt
			engine.ApplyEffects(&s, d.Effects, at)
			break
		}
		engine.SettleRound(&p, &s, rapidSettings(), d.Stake, d.Stake, at)
		at = at.Add(time.Second)
	}

	if rejectedOnBet != 6 {
		t.Fatalf("rejected on bet %d, want 6", rejectedOnBet)
	}
	if s.RapidFlags != 1 || s.RapidStreak != 0 || s.BreakUntil == nil {
		t.Errorf("session = %+v", s)
	}
}

func TestRapidFlagsRaiseFriction(t *testing.T) {
	s := openSession()
	s.RapidFlags = 2

	got := engine.FrictionLevel(plan(500, 1), settings(), &s, 0)

	if got != 2 {
		t.Errorf("friction = %d, want 2", got)
	}
}
