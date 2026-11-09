package mysql

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
)

func demoSettings(tenantID int64) model.RuleSettings {
	return model.RuleSettings{
		TenantID: tenantID, MaxStake: 50, StopLoss: 200, SessionMinutes: 180, WeeklySessions: 20,
		CooldownMinutes: 0, AllowedFromHour: 0, AllowedToHour: 24, ChasingBreakSeconds: 30,
		ConfirmEveryBets: 5, FrictionWaitSeconds: 10, WeeksPerFrictionLevel: 2, BetsPerFrictionLevel: 25,
	}
}

// Seed inserts the synthetic demo tenants when the database is empty.
func Seed(db *gorm.DB, now time.Time) error {
	var tenants int64
	err := db.Model(&model.Tenant{}).Count(&tenants).Error
	if err != nil {
		return fmt.Errorf("count tenants: %w", err)
	}
	if tenants > 0 {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		return seedDemoData(tx, now)
	})
}

func seedDemoData(tx *gorm.DB, now time.Time) error {
	weekStart := engine.WeekStartOf(now)
	pairA := model.Tenant{Name: "Pora A"}
	pairB := model.Tenant{Name: "Pora B"}
	err := tx.Create([]*model.Tenant{&pairA, &pairB}).Error
	if err != nil {
		return fmt.Errorf("create tenants: %w", err)
	}

	users := []model.User{
		{TenantID: pairA.ID, Name: "Jonas", Role: model.RoleGambler},
		{TenantID: pairA.ID, Name: "Rasa", Role: model.RoleSupervisor},
		{TenantID: pairB.ID, Name: "Tomas", Role: model.RoleGambler},
		{TenantID: pairB.ID, Name: "Eglė", Role: model.RoleSupervisor},
	}
	settings := []model.RuleSettings{demoSettings(pairA.ID), demoSettings(pairB.ID)}
	plans := []model.Plan{
		{TenantID: pairA.ID, InitialAllocation: 500, Allocation: 500, Balance: 500, WeekNumber: 1, WeekStart: weekStart},
		{TenantID: pairB.ID, InitialAllocation: 500, Allocation: 290, Balance: 250, WeekNumber: 9, WeekStart: weekStart},
	}
	chasingSession := model.Session{
		TenantID: pairB.ID, StartedAt: now.Add(-5 * time.Minute), NetLoss: 40, BetCount: 5,
		ChasingSignals: 2, LastStake: 20, LastLost: true,
	}

	for _, value := range []any{&users, &settings, &plans, &chasingSession} {
		err = tx.Create(value).Error
		if err != nil {
			return fmt.Errorf("insert seed data: %w", err)
		}
	}
	return nil
}
