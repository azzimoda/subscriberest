package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/azzimoda/subscriberest/internal/model"
	"github.com/azzimoda/subscriberest/internal/repository"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func NewService(repo repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

type SubscriptionService struct {
	repo repository.SubscriptionRepository
}

func (s *SubscriptionService) Create(c context.Context, sub *model.Subscription) error {
	return s.repo.Create(c, sub)
}

var ErrNotFound = errors.New("not found")

func (s *SubscriptionService) GetByID(c context.Context, id uuid.UUID) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(c, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
	}
	return sub, err
}

func (s *SubscriptionService) Update(c context.Context, sub *model.Subscription) error {
	return s.repo.Update(c, sub)
}

func (s *SubscriptionService) Delete(c context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(c, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		return err
	}
	return nil
}

func (s *SubscriptionService) List(c context.Context, limit, offset int) ([]model.Subscription, error) {
	return s.repo.List(c, limit, offset)
}

func (s *SubscriptionService) GetTotalPriceByUserNamePeriod(
	ctx context.Context,
	userID uuid.UUID,
	serviceName string,
	startDate, endDate time.Time,
) (int, error) {
	subs, err := s.repo.GetAllByUserIDServiceNamePeriod(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		return 0, err
	}
	log.Trace().Any("subscriptions", subs).Msg("Got all subscriptions by filters")

	total := 0
	d := startDate
	for d.Before(endDate) || d.Equal(endDate) {
		for _, sub := range subs {
			if sub.IsActiveAt(d) {
				total += sub.Price
			}
		}

		d = d.AddDate(0, 1, 0)
	}
	log.Trace().Int("total", total).Msg("Collected prices")

	return total, nil
}
