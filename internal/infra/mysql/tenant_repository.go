package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type tenantRepository struct {
	db *gorm.DB
}

func (r tenantRepository) List(ctx context.Context) ([]model.Tenant, error) {
	var tenants []model.Tenant
	err := r.db.WithContext(ctx).Order("id").Find(&tenants).Error
	return tenants, err
}
