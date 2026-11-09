//go:build integration

package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

func TestRestartIsBlockedByHistory(t *testing.T) {
	tests := []struct {
		name       string
		endReason  model.Reason
		breakFor   time.Duration
		wantReason model.Reason
	}{
		{"running break blocks a new session", model.ReasonUserEnded, time.Minute, model.ReasonBreakActive},
		{"chasing stop locks the week", model.ReasonChasingLimit, 0, model.ReasonChasingLockout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newFixture(t)
			mine := createPair(t, f.db, 500, f.now)
			updates := map[string]any{"ended_at": f.now, "end_reason": tt.endReason}
			if tt.breakFor > 0 {
				updates["break_until"] = f.now.Add(tt.breakFor)
			}
			must(t, f.db.Model(&model.Session{}).Where("id = ?", mine.sessionID).Updates(updates).Error)

			response := f.call(http.MethodPost, fmt.Sprintf("/api/tenants/%d/sessions", mine.tenantID), mine.tenantID, "")

			if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), string(tt.wantReason)) {
				t.Errorf("status = %d, body = %s, want 409 with %s", response.Code, response.Body, tt.wantReason)
			}
		})
	}
}

func TestWeeklySignalsAreSummedPerTenant(t *testing.T) {
	f := newFixture(t)
	mine := createPair(t, f.db, 500, f.now)
	theirs := createPair(t, f.db, 500, f.now)
	must(t, f.db.Model(&model.Session{}).Where("id = ?", mine.sessionID).Update("chasing_signals", 2).Error)
	must(t, f.db.Model(&model.Session{}).Where("id = ?", theirs.sessionID).Update("chasing_signals", 1).Error)

	response := f.call(http.MethodGet, fmt.Sprintf("/api/tenants/%d/state", mine.tenantID), mine.tenantID, "")

	var state struct {
		ChasingSignalsThisWeek int `json:"chasingSignalsThisWeek"`
	}
	must(t, json.Unmarshal(response.Body.Bytes(), &state))
	if state.ChasingSignalsThisWeek != 2 {
		t.Errorf("signals = %d, want 2", state.ChasingSignalsThisWeek)
	}
}
