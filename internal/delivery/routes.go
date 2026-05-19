package delivery

import (
	"iot-golang/pkg/httpclient"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB, targetURL string) Service {
	repository := NewRepository(db)
	service := NewService(repository, targetURL, httpclient.New())
	controller := NewController(service)

	router.GET("/api/v1/delivery-logs", controller.FindAllLogs)

	group := router.Group("/api/v1/deliveries")

	group.POST("/send-pending", controller.SendPending)

	return service
}
