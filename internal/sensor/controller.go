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

// Create godoc
// @Summary Create sensor
// @Description Create a new sensor record.
// @Tags Sensors
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Sensor payload"
// @Success 201 {object} DetailResponseEnvelope
// @Failure 400 {object} MessageResponse
// @Failure 409 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /sensors [post]
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

// FindAll godoc
// @Summary List sensors
// @Description Retrieve all sensors.
// @Tags Sensors
// @Accept json
// @Produce json
// @Success 200 {object} ListResponseEnvelope
// @Failure 500 {object} MessageResponse
// @Router /sensors [get]
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

// FindByID godoc
// @Summary Get sensor by ID
// @Description Retrieve a single sensor by its ID.
// @Tags Sensors
// @Accept json
// @Produce json
// @Param id path int true "Sensor ID"
// @Success 200 {object} DetailResponseEnvelope
// @Failure 404 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /sensors/{id} [get]
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

// Update godoc
// @Summary Update sensor
// @Description Update an existing sensor by its ID.
// @Tags Sensors
// @Accept json
// @Produce json
// @Param id path int true "Sensor ID"
// @Param request body UpdateRequest true "Sensor payload"
// @Success 200 {object} DetailResponseEnvelope
// @Failure 400 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Failure 409 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /sensors/{id} [put]
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

// Delete godoc
// @Summary Delete sensor
// @Description Delete a sensor by its ID.
// @Tags Sensors
// @Accept json
// @Produce json
// @Param id path int true "Sensor ID"
// @Success 200 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /sensors/{id} [delete]
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
