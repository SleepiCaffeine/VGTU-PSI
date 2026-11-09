package mysql

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type sessionRepository struct {
	db *gorm.DB
}

func (r sessionRepository) Get(ctx context.Context, tenantID, sessionID int64) (model.Session, error) {
	var session model.Session
	err := r.db.WithContext(ctx).Scopes(inTenant(tenantID)).Where("id = ?", sessionID).Take(&session).Error
	return session, translate(err)
}

func (r sessionRepository) FindOpen(ctx context.Context, tenantID int64) (*model.Session, error) {
	var sessions []model.Session
	err := r.db.WithContext(ctx).
		Scopes(inTenant(tenantID)).
		Where("ended_at IS NULL").
		Order("started_at DESC").
		Limit(1).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return firstSession(sessions), nil
}

func (r sessionRepository) Create(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r sessionRepository) Save(ctx context.Context, session model.Session) error {
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Scopes(inTenant(session.TenantID)).
		Where("id = ?", session.ID).
		Select("*").
		Updates(session).Error
}

func (r sessionRepository) CountStartedSince(ctx context.Context, tenantID int64, since time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Scopes(inTenant(tenantID)).
		Where("started_at >= ?", since).
		Count(&count).Error
	return int(count), err
}

func (r sessionRepository) LastEndedAt(ctx context.Context, tenantID int64) (*time.Time, error) {
	var sessions []model.Session
	err := r.db.WithContext(ctx).
		Scopes(inTenant(tenantID)).
		Where("ended_at IS NOT NULL").
		Order("ended_at DESC").
		Limit(1).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	last := firstSession(sessions)
	if last == nil {
		return nil, nil
	}
	return last.EndedAt, nil
}

func (r sessionRepository) CountEndedBetween(ctx context.Context, tenantID int64, reason model.Reason, from, to time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Scopes(inTenant(tenantID)).
		Where("end_reason = ? AND ended_at >= ? AND ended_at < ?", reason, from, to).
		Count(&count).Error
	return int(count), err
}

func (r sessionRepository) ShiftBack(ctx context.Context, tenantID int64, by time.Duration) error {
	seconds := int64(by.Seconds())
	return r.db.WithContext(ctx).
		Model(&model.Session{}).
		Scopes(inTenant(tenantID)).
		Updates(map[string]any{
			"started_at":  gorm.Expr("DATE_SUB(started_at, INTERVAL ? SECOND)", seconds),
			"ended_at":    gorm.Expr("DATE_SUB(ended_at, INTERVAL ? SECOND)", seconds),
			"break_until": gorm.Expr("DATE_SUB(break_until, INTERVAL ? SECOND)", seconds),
		}).Error
}

func (r sessionRepository) SumChasingSignalsSince(ctx context.Context, tenantID int64, since time.Time) (int, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Session{}).
		Scopes(inTenant(tenantID)).
		Where("started_at >= ?", since).
		Select("COALESCE(SUM(chasing_signals), 0)").
		Scan(&total).Error
	return int(total), err
}

func (r sessionRepository) LatestBreakUntil(ctx context.Context, tenantID int64) (*time.Time, error) {
	var sessions []model.Session
	err := r.db.WithContext(ctx).
		Scopes(inTenant(tenantID)).
		Where("break_until IS NOT NULL").
		Order("break_until DESC").
		Limit(1).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	latest := firstSession(sessions)
	if latest == nil {
		return nil, nil
	}
	return latest.BreakUntil, nil
}

func firstSession(sessions []model.Session) *model.Session {
	if len(sessions) == 0 {
		return nil
	}
	return &sessions[0]
}
