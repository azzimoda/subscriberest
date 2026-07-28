package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/azzimoda/subscriberest/internal/model"
	service "github.com/azzimoda/subscriberest/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func NewHandler(service *service.SubscriptionService) *Handler { return &Handler{service: service} }

type Handler struct{ service *service.SubscriptionService }

type CreateSubscriptionRequest struct {
	ServiceName string    `json:"service_name" example:"Netflix"`
	Price       int       `json:"price" example:"999" minimum:"0"`
	UserID      uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	StartDate   string    `json:"start_date" example:"2026-01"`
	EndDate     *string   `json:"end_date,omitempty" example:"2026-12"`
}

type UpdateSubscriptionRequest struct {
	ServiceName *string    `json:"service_name,omitempty" example:"Netflix"`
	Price       *int       `json:"price,omitempty" example:"999" minimum:"0"`
	UserID      *uuid.UUID `json:"user_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	StartDate   *string    `json:"start_date,omitempty" example:"2026-01"`
	EndDate     *string    `json:"end_date,omitempty" example:"2026-12"`
}

type CreateSubscriptionResponse struct {
	ID      string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message string `json:"message" example:"created"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid subscription ID format, expected UUID"`
}

type PaginationMeta struct {
	Total  int `json:"total" example:"10"`
	Limit  int `json:"limit" example:"10"`
	Offset int `json:"offset" example:"0"`
}

type ListSubscriptionsResponse struct {
	Data []model.Subscription `json:"data"`
	Meta PaginationMeta       `json:"meta"`
}

type StatsMeta struct {
	UserID      string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ServiceName string `json:"service_name" example:"Netflix"`
	StartDate   string `json:"start_date" example:"2026-01"`
	EndDate     string `json:"end_date" example:"2026-06"`
}

type GetStatsResponse struct {
	Result int       `json:"result" example:"2997"`
	Meta   StatsMeta `json:"meta"`
}

// @Summary      Create a subscription
// @Description  Create a new subscription for a user
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        request body CreateSubscriptionRequest true "Subscription data"
// @Success      201  {object}  CreateSubscriptionResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /subscriptions [post]
func (h *Handler) CreateSubscription(c *gin.Context) {
	ctx := c.Request.Context()

	var req struct {
		ServiceName string    `json:"service_name" binding:"required"`
		Price       int       `json:"price" binding:"required,min=0"`
		UserID      uuid.UUID `json:"user_id" binding:"required,uuid"`
		StartDate   string    `json:"start_date" binding:"required"`
		EndDate     *string   `json:"end_date,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Failed to parse JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := model.ParseSubscriptionDate(req.StartDate)
	if err != nil {
		log.Warn().Err(err).Str("start_date", req.StartDate).Msg("Invalid start_date value")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid start_date value: %v", err)})
		return
	}
	var endDate *time.Time = nil
	if req.EndDate != nil {
		d, err := model.ParseSubscriptionDate(*req.EndDate)
		if err != nil {
			log.Warn().Err(err).Str("end_date", *req.EndDate).Msg("Invalid end_date value")
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid end_date value: %v", err)})
			return
		}
		endDate = &d
	}

	sub := model.Subscription{
		Service:   req.ServiceName,
		Price:     req.Price,
		UserID:    req.UserID,
		StartDate: startDate,
		EndDate:   endDate,
	}
	h.service.Create(ctx, &sub)

	log.Info().Any("subscription", sub).Msg("Created subscription")
	c.JSON(http.StatusCreated, gin.H{"id": sub.ID.String(), "message": "created"})
}

// @Summary      Get subscription by ID
// @Description  Returns a subscription by its UUID
// @Tags         subscriptions
// @Produce      json
// @Param        id  path      string  true  "Subscription ID (UUID)"
// @Success      200  {object}  model.Subscription
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /subscriptions/{id} [get]
func (h *Handler) GetSubscription(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid UUID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID format, expected UUID"})
		return
	}

	subscription, err := h.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			log.Warn().Err(err).Str("uuid", id.String()).Msg("Subscription not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}

		log.Error().Err(err).Str("uuid", id.String()).Msg("Internal error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	log.Debug().Str("uuid", id.String()).Msg("Got subscription")
	c.JSON(http.StatusOK, subscription)
}

// @Summary      Update a subscription
// @Description  Update an existing subscription by ID. All fields are optional — only provided fields will be updated.
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        id       path   string                     true  "Subscription ID (UUID)"
// @Param        request  body   UpdateSubscriptionRequest  true  "Updated subscription data"
// @Success      200  {object}  model.Subscription
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /subscriptions/{id} [put]
func (h *Handler) UpdateSubscription(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid UUID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID format, expected UUID"})
		return
	}

	var req struct {
		ServiceName *string    `json:"service_name,omitempty"`
		Price       *int       `json:"price,omitempty" binding:"min=0"`
		UserID      *uuid.UUID `json:"user_id,omitempty"`
		StartDate   *string    `json:"start_date,omitempty"`
		EndDate     *string    `json:"end_date,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn().Err(err).Msg("Failed to parse JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.service.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			log.Warn().Err(err).Str("uuid", id.String()).Msg("Subscription not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}

		log.Error().Err(err).Msg("Internal error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if req.ServiceName != nil {
		existing.Service = *req.ServiceName
	}
	if req.Price != nil {
		existing.Price = *req.Price
	}
	if req.UserID != nil {
		existing.UserID = *req.UserID
	}
	if req.StartDate != nil {
		startDate, err := model.ParseSubscriptionDate(*req.StartDate)
		if err != nil {
			log.Warn().Err(err).Str("start_date", *req.StartDate).Msg("Invalid start_date value")
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid start_date value: %v", err)})
			return
		}
		existing.StartDate = startDate
	}
	if req.EndDate != nil {
		var endDate *time.Time = nil
		if *req.EndDate == "" {
			endDate = nil
		} else {
			d, err := model.ParseSubscriptionDate(*req.EndDate)
			if err != nil {
				log.Warn().Err(err).Str("end_date", *req.EndDate).Msg("Invalid end_date value")
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid end_date value: %v", err)})
				return
			}
			endDate = &d
		}
		existing.EndDate = endDate
	}

	if err := h.service.Update(ctx, existing); err != nil {
		log.Error().Err(err).Msg("Internal error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	log.Info().Str("uuid", id.String()).Any("request", req).Msg("Updated subscription")
	c.JSON(http.StatusOK, existing)
}

// @Summary      Delete a subscription
// @Description  Delete a subscription by ID
// @Tags         subscriptions
// @Produce      json
// @Param        id  path  string  true  "Subscription ID (UUID)"
// @Success      204  "No Content"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /subscriptions/{id} [delete]
func (h *Handler) DeleteSubscription(c *gin.Context) {
	ctx := c.Request.Context()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid UUID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription ID format, expected UUID"})
		return
	}

	if _, err := h.service.GetByID(ctx, id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			log.Warn().Err(err).Str("uuid", id.String()).Msg("Subscription not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			return
		}

		log.Error().Err(err).Str("uuid", id.String()).Msg("Internal error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if err := h.service.Delete(ctx, id); err != nil {
		log.Error().Err(err).Str("uuid", id.String()).Msg("Internal error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	log.Info().Str("uuid", id.String()).Msg("Deleted subscription")
	c.Status(http.StatusNoContent)
}

// @Summary      List subscriptions
// @Description  Returns a paginated list of all subscriptions
// @Tags         subscriptions
// @Produce      json
// @Param        limit  query  int  false  "Number of rows per page"  default(10)
// @Param        offset query  int  false  "Offset for pagination"    default(0)
// @Success      200  {object}  ListSubscriptionsResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /subscriptions [get]
func (h *Handler) ListSubscriptions(c *gin.Context) {
	ctx := c.Request.Context()

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	subscriptions, err := h.service.List(ctx, limit, offset)
	if err != nil {
		log.Error().Err(err).Msg("Internal error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	total := len(subscriptions)

	log.Debug().Int("total", total).Int("limit", limit).Int("offset", offset).Msg("Listed subscriptions")
	c.JSON(http.StatusOK, gin.H{
		"data": subscriptions,
		"meta": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// @Summary      Get subscription stats
// @Description  Calculate total cost of subscriptions for a user within a date range, optionally filtered by service name
// @Tags         subscriptions
// @Produce      json
// @Param        user_id      query  string  true   "User ID (UUID)"
// @Param        service_name query  string  false  "Filter by service name (partial match)"
// @Param        start_date   query  string  true   "Start date (YYYY-MM)"
// @Param        end_date     query  string  true   "End date (YYYY-MM)"
// @Success      200  {object}  GetStatsResponse
// @Failure      400  {object}  ErrorResponse
// @Router       /subscriptions/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	ctx := c.Request.Context()

	// Get query values

	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		log.Warn().Msg("No UUID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Warn().Err(err).Str("uuid", userIDStr).Msg("Invalid UUID")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	serviceName := c.Query("service_name")

	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		log.Warn().Msg("No start date")
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date is required"})
		return
	}
	startDate, err := model.ParseSubscriptionDate(startDateStr)
	if err != nil {
		log.Warn().Err(err).Str("start_date", startDateStr).Msg("Invalid start_date value")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid start_date value: %v", err)})
		return
	}

	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		log.Warn().Msg("No end date")
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date is required"})
		return
	}
	endDate, err := model.ParseSubscriptionDate(endDateStr)
	if err != nil {
		log.Warn().Err(err).Str("end_date", endDateStr).Msg("Invalid end_date value")
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid end_date value: %v", err)})
		return
	}

	// Get result

	result, err := h.service.GetTotalPriceByUserNamePeriod(ctx, userID, serviceName, startDate, endDate)
	if err != nil {
		log.Error().Err(err).Str("uuid", userIDStr).Str("serviceName", serviceName).
			Str("start_date", startDateStr).Str("end_date", endDateStr).
			Msg("Internal error")
		c.JSON(http.StatusBadRequest, gin.H{"error": "internal server error"})
		return
	}

	log.Debug().Str("user_id", userIDStr).Str("service_name", serviceName).
		Str("start_date", startDateStr).Str("end_date", endDateStr).
		Int("result", result).
		Msg("Got stats")
	c.JSON(http.StatusOK, gin.H{
		"result": result,
		"meta": gin.H{
			"service_name": serviceName,
			"user_id":      userIDStr,
			"start_date":   startDateStr,
			"end_date":     endDateStr,
		},
	})
}
