package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type betRepository struct {
	db *gorm.DB
}

func (r betRepository) Create(ctx context.Context, bet *model.Bet) error {
	return r.db.WithContext(ctx).Create(bet).Error
}
