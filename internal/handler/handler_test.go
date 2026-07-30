package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/azzimoda/subscriberest/internal/handler"
	"github.com/azzimoda/subscriberest/internal/model"
	"github.com/azzimoda/subscriberest/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockService struct {
	mock.Mock
}

var _ handler.Service = (*mockService)(nil)

func (m *mockService) Create(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *mockService) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *mockService) Update(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *mockService) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockService) List(ctx context.Context, limit, offset int) ([]model.Subscription, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]model.Subscription), args.Error(1)
}

func (m *mockService) GetTotalPriceByUserNamePeriod(
	ctx context.Context,
	userID uuid.UUID,
	serviceName string,
	startDate, endDate time.Time,
) (int, error) {
	args := m.Called(ctx, userID, serviceName, startDate, endDate)
	return args.Int(0), args.Error(1)
}

func setupRouter(h *handler.Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/subscriptions", h.CreateSubscription)
	r.GET("/api/v1/subscriptions/:id", h.GetSubscription)
	r.PUT("/api/v1/subscriptions/:id", h.UpdateSubscription)
	r.DELETE("/api/v1/subscriptions/:id", h.DeleteSubscription)
	r.GET("/api/v1/subscriptions", h.ListSubscriptions)
	r.GET("/api/v1/subscriptions/stats", h.GetStats)
	return r
}

func TestCreateSubscription_Success(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	body := `{"service_name":"Netflix","price":999,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"2026-01"}`
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(nil)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "created", resp["message"])
	assert.NotEmpty(t, resp["id"])
	mockSvc.AssertExpectations(t)
}

func TestCreateSubscription_BadRequest(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	body := `{"price":999}`
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateSubscription_InternalError(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	body := `{"service_name":"Netflix","price":999,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"2026-01"}`
	mockSvc.On("Create", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(errors.New("db error"))

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/subscriptions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestGetSubscription_Found(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	id := uuid.New()
	sub := &model.Subscription{ID: id, Service: "Netflix", Price: 999}
	mockSvc.On("GetByID", mock.Anything, id).Return(sub, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/subscriptions/"+id.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestGetSubscription_NotFound(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	id := uuid.New()
	mockSvc.On("GetByID", mock.Anything, id).Return(nil, service.ErrNotFound)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/subscriptions/"+id.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestGetSubscription_InvalidUUID(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/subscriptions/invalid-uuid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateSubscription_Success(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	id := uuid.New()
	existing := &model.Subscription{ID: id, Service: "Netflix", Price: 999}
	mockSvc.On("GetByID", mock.Anything, id).Return(existing, nil)
	mockSvc.On("Update", mock.Anything, mock.AnythingOfType("*model.Subscription")).Return(nil)

	body := `{"price":499}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/subscriptions/"+id.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestUpdateSubscription_NotFound(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	id := uuid.New()
	mockSvc.On("GetByID", mock.Anything, id).Return(nil, service.ErrNotFound)

	body := `{"price":499}`
	req, _ := http.NewRequest(http.MethodPut, "/api/v1/subscriptions/"+id.String(), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestDeleteSubscription_Success(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	id := uuid.New()
	sub := &model.Subscription{ID: id}
	mockSvc.On("GetByID", mock.Anything, id).Return(sub, nil)
	mockSvc.On("Delete", mock.Anything, id).Return(nil)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/subscriptions/"+id.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestDeleteSubscription_NotFound(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	id := uuid.New()
	mockSvc.On("GetByID", mock.Anything, id).Return(nil, service.ErrNotFound)

	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/subscriptions/"+id.String(), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockSvc.AssertExpectations(t)
}

func TestListSubscriptions(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	subs := []model.Subscription{{Service: "Netflix"}}
	mockSvc.On("List", mock.Anything, 10, 0).Return(subs, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/subscriptions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NotNil(t, resp["data"])
	assert.NotNil(t, resp["meta"])
	mockSvc.AssertExpectations(t)
}

func TestGetStats_Success(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	userID := uuid.New()
	mockSvc.On("GetTotalPriceByUserNamePeriod", mock.Anything, userID, "Netflix",
		mock.Anything, mock.Anything).Return(2997, nil)

	req, _ := http.NewRequest(http.MethodGet,
		"/api/v1/subscriptions/stats?user_id="+userID.String()+"&service_name=Netflix&start_date=2026-01&end_date=2026-03", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(2997), resp["result"])
	mockSvc.AssertExpectations(t)
}

func TestGetStats_MissingUserID(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/subscriptions/stats?start_date=2026-01&end_date=2026-03", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetStats_InvalidUserID(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	req, _ := http.NewRequest(http.MethodGet,
		"/api/v1/subscriptions/stats?user_id=bad-uuid&start_date=2026-01&end_date=2026-03", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetStats_ServiceError(t *testing.T) {
	mockSvc := new(mockService)
	h := handler.NewHandler(mockSvc)
	r := setupRouter(h)

	userID := uuid.New()
	mockSvc.On("GetTotalPriceByUserNamePeriod", mock.Anything, userID, "",
		mock.Anything, mock.Anything).Return(0, errors.New("internal error"))

	req, _ := http.NewRequest(http.MethodGet,
		"/api/v1/subscriptions/stats?user_id="+userID.String()+"&start_date=2026-01&end_date=2026-03", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
