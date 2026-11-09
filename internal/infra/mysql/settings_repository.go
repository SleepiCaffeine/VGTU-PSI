package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type settingsRepository struct {
	db *gorm.DB
}

func (r settingsRepository) Get(ctx context.Context, tenantID int64) (model.RuleSettings, error) {
	var settings model.RuleSettings
	err := r.db.WithContext(ctx).Scopes(inTenant(tenantID)).Take(&settings).Error
	return settings, translate(err)
}
