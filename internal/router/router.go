package router

import (
	"github.com/azzimoda/subscriberest/internal/handler"
	"github.com/gin-gonic/gin"
)

func Init(h *handler.Handler) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		api.POST("/subscriptions", h.CreateSubscription)
		api.GET("/subscriptions/:id", h.GetSubscription)
		api.PUT("/subscriptions/:id", h.UpdateSubscription)
		api.DELETE("/subscriptions/:id", h.DeleteSubscription)
		api.GET("/subscriptions", h.ListSubscriptions)

		api.GET("/subscriptions/stats", h.GetStats)
	}

	return r
}
