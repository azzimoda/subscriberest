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
	GetAllByUserIDServiceNamePeriod(
		ctx context.Context,
		userID uuid.UUID,
		serviceName string,
		startDate, endDate time.Time,
	) ([]model.Subscription, error)
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

func (r *subscriptionRepository) GetAllByUserIDServiceNamePeriod(
	ctx context.Context,
	userID uuid.UUID,
	serviceName string,
	startDate, endDate time.Time,
) ([]model.Subscription, error) {
	var subs []model.Subscription
	q := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("start_date <= ?", endDate).
		Where("(end_date IS NULL OR end_date >= ?)", startDate)
	if serviceName != "" {
		q = q.Where("service ILIKE ?", "%"+serviceName+"%")
	}
	tx := q.Order("start_date DESC").Find(&subs)
	return subs, tx.Error
}
