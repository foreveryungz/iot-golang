package monitoring

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

// Create godoc
// @Summary Create monitoring data
// @Description Create a monitoring data record for a device and sensor.
// @Tags Monitoring
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Monitoring payload"
// @Success 201 {object} DetailResponseEnvelope
// @Failure 400 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /monitoring-data [post]
func (h *Controller) Create(c *gin.Context) {
	var input CreateRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	result, err := h.service.Create(&input)
	if err != nil {
		switch {
		case errors.Is(err, ErrDeviceNotFound), errors.Is(err, ErrSensorNotFound):
			c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
		case errors.Is(err, ErrSensorDeviceMismatch):
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": result})
}

// FindAll godoc
// @Summary List monitoring data
// @Description Retrieve all monitoring data records.
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} ListResponseEnvelope
// @Failure 500 {object} MessageResponse
// @Router /monitoring-data [get]
func (h *Controller) FindAll(c *gin.Context) {
	items, err := h.service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}
