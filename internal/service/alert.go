package service

import (
	"context"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type AlertService struct {
	deps Dependencies
}

func NewAlertService(deps Dependencies) *AlertService {
	return &AlertService{deps: deps}
}

func (s *AlertService) List(ctx context.Context, tenantID int64) ([]model.Alert, error) {
	var alerts []model.Alert
	err := s.deps.UnitOfWork.Do(ctx, func(ctx context.Context, repos Repositories) error {
		var err error
		alerts, err = repos.Alerts.List(ctx, tenantID)
		return err
	})
	return alerts, err
}
