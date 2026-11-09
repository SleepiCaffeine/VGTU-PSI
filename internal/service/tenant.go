package service

import (
	"context"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

type TenantService struct {
	deps Dependencies
}

func NewTenantService(deps Dependencies) *TenantService {
	return &TenantService{deps: deps}
}

func (s *TenantService) List(ctx context.Context) ([]model.Tenant, error) {
	var tenants []model.Tenant
	err := s.deps.UnitOfWork.Do(ctx, func(ctx context.Context, repos Repositories) error {
		var err error
		tenants, err = repos.Tenants.List(ctx)
		return err
	})
	return tenants, err
}
