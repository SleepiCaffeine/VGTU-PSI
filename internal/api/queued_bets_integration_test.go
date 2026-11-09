//go:build integration

package api_test

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

func TestQueuedBetsStillHitRules(t *testing.T) {
	const queuedBets = 100

	tests := []struct {
		name          string
		rapidBetLimit int
		wantAccepted  int64
	}{
		{"rapid betting stops the queue", 5, 5},
		{"friction stops the queue when the rapid rule is off", 0, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			mine := createPair(t, f.db, 500, f.now)
			must(t, f.db.Model(&model.RuleSettings{}).Where("tenant_id = ?", mine.tenantID).Updates(map[string]any{
				"bets_per_friction_level": 5, "rapid_bet_seconds": 3, "rapid_bet_limit": tt.rapidBetLimit,
			}).Error)
			path := fmt.Sprintf("/api/tenants/%d/sessions/%d/bets", mine.tenantID, mine.sessionID)

			var accepted atomic.Int64
			var requests sync.WaitGroup
			for range queuedBets {
				requests.Go(func() {
					response := f.call(http.MethodPost, path, mine.tenantID, `{"game":"roulette","selection":"red","stake":5}`)
					if response.Code == http.StatusOK {
						accepted.Add(1)
					}
				})
			}
			requests.Wait()

			after := f.snapshot(t, mine)
			if accepted.Load() != tt.wantAccepted || after.bets != tt.wantAccepted {
				t.Errorf("accepted = %d, recorded bets = %d, want %d", accepted.Load(), after.bets, tt.wantAccepted)
			}
			if after.plan.Balance != 500-5*int(tt.wantAccepted) {
				t.Errorf("balance = %d", after.plan.Balance)
			}
		})
	}
}
