package main

import (
	"iot-golang/internal/config"
	"iot-golang/internal/database"
	"iot-golang/internal/delivery"
	"iot-golang/internal/device"
	"iot-golang/internal/monitoring"
	"iot-golang/internal/sensor"
	"log"

	"github.com/gin-gonic/gin"
)

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
