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
		CooldownMinutes: 15, AllowedFromHour: 0, AllowedToHour: 24, ChasingBreakSeconds: 30,
		ConfirmEveryBets: 5, FrictionWaitSeconds: 10, WeeksPerFrictionLevel: 2, BetsPerFrictionLevel: 5,
		RapidBetSeconds: 3, RapidBetLimit: 5,
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
	pairC := model.Tenant{Name: "Pora C"}
	err := tx.Create([]*model.Tenant{&pairA, &pairB, &pairC}).Error
	if err != nil {
		return fmt.Errorf("create tenants: %w", err)
	}

	users := []model.User{
		{TenantID: pairA.ID, Name: "Jonas", Role: model.RoleGambler},
		{TenantID: pairA.ID, Name: "Rasa", Role: model.RoleSupervisor},
		{TenantID: pairB.ID, Name: "Tomas", Role: model.RoleGambler},
		{TenantID: pairB.ID, Name: "Eglė", Role: model.RoleSupervisor},
		{TenantID: pairC.ID, Name: "Mantas", Role: model.RoleGambler},
		{TenantID: pairC.ID, Name: "Ieva", Role: model.RoleSupervisor},
	}
	settings := []model.RuleSettings{demoSettings(pairA.ID), demoSettings(pairB.ID), demoSettings(pairC.ID)}
	plans := []model.Plan{
		{TenantID: pairA.ID, InitialAllocation: 500, Allocation: 500, Balance: 500, WeekNumber: 1, WeekStart: weekStart},
		{TenantID: pairB.ID, InitialAllocation: 500, Allocation: 290, Balance: 250, WeekNumber: 9, WeekStart: weekStart},
		{TenantID: pairC.ID, InitialAllocation: 500, Allocation: 290, Balance: 250, WeekNumber: 9, WeekStart: weekStart},
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

	alerts := []model.Alert{
		{TenantID: pairB.ID, SessionID: chasingSession.ID, Reason: model.ReasonChasingSignal, CreatedAt: now.Add(-3 * time.Minute)},
		{TenantID: pairB.ID, SessionID: chasingSession.ID, Reason: model.ReasonChasingSignal, CreatedAt: now.Add(-time.Minute)},
	}
	err = tx.Create(&alerts).Error
	if err != nil {
		return fmt.Errorf("insert seed alerts: %w", err)
	}
	return nil
}
