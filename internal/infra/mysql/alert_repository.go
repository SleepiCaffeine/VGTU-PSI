package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type alertRepository struct {
	db *gorm.DB
}

func (r alertRepository) Create(ctx context.Context, alert *model.Alert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

func (r alertRepository) List(ctx context.Context, tenantID int64) ([]model.Alert, error) {
	var alerts []model.Alert
	err := r.db.WithContext(ctx).Scopes(inTenant(tenantID)).Order("id DESC").Find(&alerts).Error
	return alerts, err
}
