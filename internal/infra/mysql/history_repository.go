package mysql

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

const weekActivityColumns = `
	COUNT(CASE WHEN started_at >= @from AND started_at < @to THEN 1 END) AS sessions_started,
	COALESCE(SUM(CASE WHEN started_at >= @from AND started_at < @to THEN chasing_signals END), 0) AS chasing_signals,
	COUNT(CASE WHEN end_reason = @stop AND ended_at >= @from AND ended_at < @to THEN 1 END) AS chasing_stops,
	MAX(ended_at) AS last_session_ended_at,
	MAX(break_until) AS latest_break_until`

type historyRepository struct {
	db *gorm.DB
}

func (r historyRepository) WeekActivity(ctx context.Context, tenantID int64, from, to time.Time) (model.WeekActivity, error) {
	var activity model.WeekActivity
	err := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Scopes(inTenant(tenantID)).
		Select(weekActivityColumns, map[string]any{"from": from, "to": to, "stop": model.ReasonChasingLimit}).
		Scan(&activity).Error
	return activity, err
}
