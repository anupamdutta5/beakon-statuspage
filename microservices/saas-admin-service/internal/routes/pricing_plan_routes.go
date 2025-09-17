package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/anupamdutta5/statuspage-saas-admin-service/internal/handlers"
)

// RegisterPricingPlanRoutes registers all pricing plan related routes
func RegisterPricingPlanRoutes(router *gin.RouterGroup, handler *handlers.PricingPlanHandler) {
	pricingGroup := router.Group("/plans")
	{
		pricingGroup.GET("", handler.ListPricingPlans)
		pricingGroup.POST("", handler.CreatePricingPlan)
		pricingGroup.GET("/:id", handler.GetPricingPlan)
		pricingGroup.PUT("/:id", handler.UpdatePricingPlan)
		pricingGroup.DELETE("/:id", handler.DeletePricingPlan)
	}
}
