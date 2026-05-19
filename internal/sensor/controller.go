package sensor

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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	result, err := h.service.Create(&input)
	if err != nil {
		status := mapError(err)
		c.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    http.StatusCreated,
		"message": "Sensor created",
		"data":    result,
	})
}

func (h *Controller) FindAll(c *gin.Context) {
	sensors, err := h.service.FindAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Sensors found",
		"data":    sensors,
	})
}

func (h *Controller) FindByID(c *gin.Context) {
	sensor, err := h.service.FindByID(c.Param("id"))
	if err != nil {
		status := mapError(err)
		c.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Sensor found",
		"data":    sensor,
	})
}

func (h *Controller) Update(c *gin.Context) {
	var input UpdateRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	result, err := h.service.Update(c.Param("id"), &input)
	if err != nil {
		status := mapError(err)
		c.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Sensor updated",
		"data":    result,
	})
}

func (h *Controller) Delete(c *gin.Context) {
	if err := h.service.Delete(c.Param("id")); err != nil {
		status := mapError(err)
		c.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Sensor deleted",
		"data":    nil,
	})
}

func mapError(err error) int {
	switch {
	case errors.Is(err, ErrSensorNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrDuplicateSensorCode):
		return http.StatusConflict
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidStatus), errors.Is(err, ErrInvalidDeviceID):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
