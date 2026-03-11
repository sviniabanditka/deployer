package billing

import (
	"context"
	"fmt"

	"github.com/deployer/api/internal/model"
	"github.com/google/uuid"
)

// EnforceAppQuota checks whether the user can create a new app under their current plan.
func (s *BillingService) EnforceAppQuota(ctx context.Context, userID uuid.UUID) error {
	allowed, err := s.CheckQuota(ctx, userID, "app")
	if err != nil {
		return fmt.Errorf("failed to check app quota: %w", err)
	}
	if !allowed {
		return ErrQuotaExceeded
	}
	return nil
}

// EnforceDBQuota checks whether the user can create a new database under their current plan.
func (s *BillingService) EnforceDBQuota(ctx context.Context, userID uuid.UUID) error {
	allowed, err := s.CheckQuota(ctx, userID, "database")
	if err != nil {
		return fmt.Errorf("failed to check database quota: %w", err)
	}
	if !allowed {
		return ErrQuotaExceeded
	}
	return nil
}

// GetUsageSummary returns the current resource usage vs plan limits for a user.
func (s *BillingService) GetUsageSummary(ctx context.Context, userID uuid.UUID) (*model.UsageSummary, error) {
	plan, _, err := s.GetCurrentPlan(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current plan: %w", err)
	}

	appCount, err := s.billingRepo.CountAppsByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count apps: %w", err)
	}

	dbCount, err := s.billingRepo.CountDatabasesByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to count databases: %w", err)
	}

	storageUsed, err := s.billingRepo.SumDatabaseStorageByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate storage usage: %w", err)
	}

	return &model.UsageSummary{
		AppCount:    appCount,
		AppLimit:    plan.AppLimit,
		DBCount:     dbCount,
		DBLimit:     plan.DBLimit,
		StorageUsed: storageUsed,
		StorageMax:  plan.StorageLimit,
	}, nil
}
