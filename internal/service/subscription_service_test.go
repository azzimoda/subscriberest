package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/azzimoda/subscriberest/internal/model"
	"github.com/azzimoda/subscriberest/internal/repository"
	"github.com/azzimoda/subscriberest/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Create(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *mockRepository) Update(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRepository) List(ctx context.Context, limit, offset int) ([]model.Subscription, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]model.Subscription), args.Error(1)
}

func (m *mockRepository) GetTotalPriceByUserNamePeriod(
	ctx context.Context,
	userID uuid.UUID,
	serviceName string,
	startDate, endDate time.Time,
) (int, error) {
	args := m.Called(ctx, userID, serviceName, startDate, endDate)
	return args.Int(0), args.Error(1)
}

var _ repository.SubscriptionRepository = (*mockRepository)(nil)

func TestCreate(t *testing.T) {
	mockRepo := new(mockRepository)
	s := service.NewService(mockRepo)

	sub := &model.Subscription{Service: "Netflix", Price: 999}
	mockRepo.On("Create", mock.Anything, sub).Return(nil)

	err := s.Create(context.Background(), sub)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_Found(t *testing.T) {
	mockRepo := new(mockRepository)
	s := service.NewService(mockRepo)

	id := uuid.New()
	expected := &model.Subscription{ID: id, Service: "Netflix"}
	mockRepo.On("GetByID", mock.Anything, id).Return(expected, nil)

	got, err := s.GetByID(context.Background(), id)

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	mockRepo.AssertExpectations(t)
}

func TestGetByID_NotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	s := service.NewService(mockRepo)

	id := uuid.New()
	mockRepo.On("GetByID", mock.Anything, id).Return(nil, gorm.ErrRecordNotFound)

	_, err := s.GetByID(context.Background(), id)

	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestUpdate(t *testing.T) {
	mockRepo := new(mockRepository)
	s := service.NewService(mockRepo)

	sub := &model.Subscription{ID: uuid.New(), Service: "Netflix"}
	mockRepo.On("Update", mock.Anything, sub).Return(nil)

	err := s.Update(context.Background(), sub)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDelete_OK(t *testing.T) {
	mockRepo := new(mockRepository)
	s := service.NewService(mockRepo)

	id := uuid.New()
	mockRepo.On("Delete", mock.Anything, id).Return(nil)

	err := s.Delete(context.Background(), id)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	mockRepo := new(mockRepository)
	s := service.NewService(mockRepo)

	id := uuid.New()
	mockRepo.On("Delete", mock.Anything, id).Return(gorm.ErrRecordNotFound)

	err := s.Delete(context.Background(), id)

	assert.Error(t, err)
	assert.ErrorIs(t, err, service.ErrNotFound)
	mockRepo.AssertExpectations(t)
}

func TestList(t *testing.T) {
	mockRepo := new(mockRepository)
	s := service.NewService(mockRepo)

	expected := []model.Subscription{{Service: "Netflix"}}
	mockRepo.On("List", mock.Anything, 10, 0).Return(expected, nil)

	got, err := s.List(context.Background(), 10, 0)

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	mockRepo.AssertExpectations(t)
}

func TestGetTotalPriceByUserNamePeriod(t *testing.T) {
	mockRepo := new(mockRepository)
	s := service.NewService(mockRepo)

	userID := uuid.New()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	mockRepo.On("GetTotalPriceByUserNamePeriod", mock.Anything, userID, "Netflix", start, end).Return(2997, nil)

	total, err := s.GetTotalPriceByUserNamePeriod(context.Background(), userID, "Netflix", start, end)

	assert.NoError(t, err)
	assert.Equal(t, 2997, total)
	mockRepo.AssertExpectations(t)
}

func TestGetTotalPriceByUserNamePeriod_Error(t *testing.T) {
	mockRepo := new(mockRepository)
	s := service.NewService(mockRepo)

	userID := uuid.New()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	mockRepo.On("GetTotalPriceByUserNamePeriod", mock.Anything, userID, "", start, end).Return(0, errors.New("db error"))

	_, err := s.GetTotalPriceByUserNamePeriod(context.Background(), userID, "", start, end)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
