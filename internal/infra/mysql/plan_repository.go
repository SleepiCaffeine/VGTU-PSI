package mysql

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type planRepository struct {
	db *gorm.DB
}

func (r planRepository) GetForUpdate(ctx context.Context, tenantID int64) (model.Plan, error) {
	var plan model.Plan
	err := r.db.WithContext(ctx).
		Scopes(inTenant(tenantID)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Take(&plan).Error
	return plan, translate(err)
}

func (r planRepository) Save(ctx context.Context, plan model.Plan) error {
	return r.db.WithContext(ctx).
		Model(&model.Plan{}).
		Scopes(inTenant(plan.TenantID)).
		Where("id = ?", plan.ID).
		Select("*").
		Updates(plan).Error
}
