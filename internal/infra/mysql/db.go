// Package mysql stores application data in MySQL through GORM.
package mysql

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service"
)

// DefaultDSN matches the database defined in docker-compose.yml.
const DefaultDSN = "nuolydis:nuolydis@tcp(127.0.0.1:3307)/nuolydis?charset=utf8mb4&parseTime=true&loc=UTC"

func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return db, nil
}

func Migrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&model.Tenant{}, &model.User{}, &model.RuleSettings{}, &model.Plan{},
		&model.Session{}, &model.Bet{}, &model.Confirmation{}, &model.Alert{},
	)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

// inTenant is the single place that restricts a query to one tenant.
func inTenant(tenantID int64) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}
}

func translate(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return service.ErrNotFound
	}
	return err
}

// UnitOfWork runs service work inside a database transaction.
type UnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, repos service.Repositories) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, newRepositories(tx))
	})
}

func newRepositories(tx *gorm.DB) service.Repositories {
	return service.Repositories{
		Tenants:       tenantRepository{db: tx},
		Plans:         planRepository{db: tx},
		Settings:      settingsRepository{db: tx},
		Sessions:      sessionRepository{db: tx},
		History:       historyRepository{db: tx},
		Bets:          betRepository{db: tx},
		Confirmations: confirmationRepository{db: tx},
		Alerts:        alertRepository{db: tx},
	}
}
