package device

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	repo := NewRepository(db)
	service := NewService(repo)
	controller := NewController(service)

	group := router.Group("/api/v1/devices")

	group.POST("", controller.Create)
	group.GET("", controller.FindAll)
	group.GET("/:id", controller.FindByID)
	group.PUT("/:id", controller.Update)
	group.DELETE("/:id", controller.Delete)
}
