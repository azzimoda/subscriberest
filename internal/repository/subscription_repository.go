package repository

import (
	"context"
	"time"

	"github.com/azzimoda/subscriberest/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

type SubscriptionRepository interface {
	Create(context.Context, *model.Subscription) error
	GetByID(context.Context, uuid.UUID) (*model.Subscription, error)
	Update(context.Context, *model.Subscription) error
	Delete(context.Context, uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]model.Subscription, error)
	GetTotalPriceByUserNamePeriod(
		ctx context.Context,
		userID uuid.UUID,
		serviceName string,
		startDate, endDate time.Time,
	) (int, error)
}

type subscriptionRepository struct{ db *gorm.DB }

func (r *subscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	tx := r.db.WithContext(ctx).Create(sub)
	return tx.Commit().Error
}

func (r *subscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	var sub model.Subscription
	tx := r.db.WithContext(ctx).Model(model.Subscription{}).First(&sub, "id = ?", id)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &sub, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	tx := r.db.WithContext(ctx).Where("id = ?", sub.ID).Updates(sub)
	if tx.Error != nil {
		return tx.Error
	}

	if sub.EndDate == nil {
		tx := r.db.WithContext(ctx).Model(model.Subscription{}).Where("id = ?", sub.ID).Update("end_date", nil)
		return tx.Error
	}
	return nil
}

func (r *subscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx := r.db.WithContext(ctx).Delete(&model.Subscription{}, id)
	return tx.Error
}

func (r *subscriptionRepository) List(ctx context.Context, limit, offset int) ([]model.Subscription, error) {
	var subs []model.Subscription
	tx := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&subs)
	return subs, tx.Error
}

func (r *subscriptionRepository) GetTotalPriceByUserNamePeriod(
	ctx context.Context,
	userID uuid.UUID,
	serviceName string,
	startDate, endDate time.Time,
) (int, error) {
	var total int
	q := `SELECT COALESCE(SUM(s.price), 0) AS total
		FROM subscriptions s
		CROSS JOIN generate_series(?::date, ?::date, '1 month'::interval) AS g(m)
		WHERE s.user_id = ?
			AND (? = '' OR s.service ILIKE '%' || ? || '%')
			AND s.start_date <= g.m::date
			AND (s.end_date IS NULL OR s.end_date >= g.m::date)`
	tx := r.db.WithContext(ctx).Raw(q, startDate, endDate, userID, serviceName, serviceName).Scan(&total)
	return total, tx.Error
}
