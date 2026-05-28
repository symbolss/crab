package services

import (
	"context"
	"time"

	"github.com/family/crab-server/internal/models"
	"github.com/family/crab-server/internal/repository"
	"github.com/google/uuid"
)

// PlayEventService handles play event business logic
type PlayEventService struct {
	childRepo repository.ChildRepository
	eventRepo repository.PlayEventRepository
	usageRepo repository.DailyUsageRepository
}

// NewPlayEventService creates a new play event service
func NewPlayEventService(
	childRepo repository.ChildRepository,
	eventRepo repository.PlayEventRepository,
	usageRepo repository.DailyUsageRepository,
) *PlayEventService {
	return &PlayEventService{
		childRepo: childRepo,
		eventRepo: eventRepo,
		usageRepo: usageRepo,
	}
}

// ReportEvent records a play event and updates daily usage
func (s *PlayEventService) ReportEvent(ctx context.Context, childID uuid.UUID, event *models.PlayEvent) error {
	// Check daily usage limit
	child, err := s.childRepo.GetByID(ctx, childID)
	if err != nil {
		return err
	}

	// Get today's usage
	today := time.Now().Truncate(24 * time.Hour)
	usage, err := s.usageRepo.GetByChildAndDate(ctx, childID, today)
	if err != nil {
		// Create new usage record if not exists
		usage = &models.DailyUsage{
			ChildID:      childID,
			Date:         today,
			TotalPlaySec: 0,
		}
	}

	// Check if usage exceeded (only for start/complete events)
	if event.EventType == "start" || event.EventType == "complete" {
		if usage.TotalPlaySec >= child.DailyTimeLimitSec {
			return models.ErrUsageExceeded
		}
	}

	// Record the event
	if err := s.eventRepo.Create(ctx, childID, event); err != nil {
		return err
	}

	// Update daily usage for start events
	if event.EventType == "start" && event.PositionSec > 0 {
		usage.TotalPlaySec += event.PositionSec
		if err := s.usageRepo.Upsert(ctx, usage); err != nil {
			return err
		}
	}

	return nil
}
