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

func (h *Controller) FindAll(c *gin.Context) {
	items, err := h.service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": items})
}
