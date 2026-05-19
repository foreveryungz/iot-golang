package monitoring

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	repository := NewRepository(db)
	service := NewService(repository)
	controller := NewController(service)

	group := router.Group("/api/v1/monitoring-data")

	group.POST("", controller.Create)
	group.GET("", controller.FindAll)
}
