package main

import (
	_ "iot-golang/docs"
	"iot-golang/internal/config"
	"iot-golang/internal/database"
	"iot-golang/internal/delivery"
	"iot-golang/internal/device"
	"iot-golang/internal/monitoring"
	"iot-golang/internal/sensor"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title IoT Device Delivery API
// @version 1.0
// @description Backend API for IoT monitoring data delivery with retry and dead-letter handling.
// @host localhost:8080
// @BasePath /api/v1
func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "API is running..",
		})
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	device.RegisterRoutes(router, db)
	sensor.RegisterRoutes(router, db)
	monitoring.RegisterRoutes(router, db)
	deliveryService := delivery.RegisterRoutes(router, db, cfg.ClientURL)

	if err := delivery.StartScheduler(deliveryService); err != nil {
		log.Fatal("Failed to start delivery scheduler: ", err)
	}

	log.Println("Server is running on port " + cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
