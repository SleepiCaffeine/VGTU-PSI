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
	if len(sessions) == 0 {
		return nil, nil
	}
	return &sessions[0], nil
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
