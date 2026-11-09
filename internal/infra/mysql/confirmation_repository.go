package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type confirmationRepository struct {
	db *gorm.DB
}

func (r confirmationRepository) Find(ctx context.Context, tenantID, confirmationID int64) (*model.Confirmation, error) {
	var confirmations []model.Confirmation
	err := r.db.WithContext(ctx).
		Scopes(inTenant(tenantID)).
		Where("id = ?", confirmationID).
		Limit(1).
		Find(&confirmations).Error
	if err != nil {
		return nil, err
	}
	if len(confirmations) == 0 {
		return nil, nil
	}
	return &confirmations[0], nil
}

func (r confirmationRepository) Create(ctx context.Context, confirmation *model.Confirmation) error {
	return r.db.WithContext(ctx).Create(confirmation).Error
}

func (r confirmationRepository) Save(ctx context.Context, confirmation model.Confirmation) error {
	return r.db.WithContext(ctx).
		Model(&model.Confirmation{}).
		Scopes(inTenant(confirmation.TenantID)).
		Where("id = ?", confirmation.ID).
		Select("*").
		Updates(confirmation).Error
}
