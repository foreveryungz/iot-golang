package device

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
// @Summary Create device
// @Description Create a new device record.
// @Tags Devices
// @Accept json
// @Produce json
// @Param request body CreateRequest true "Device payload"
// @Success 201 {object} DetailResponseEnvelope
// @Failure 400 {object} MessageResponse
// @Failure 409 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /devices [post]
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
		"message": "Device created",
		"data":    result,
	})
}

// FindAll godoc
// @Summary List devices
// @Description Retrieve all devices.
// @Tags Devices
// @Accept json
// @Produce json
// @Success 200 {object} ListResponseEnvelope
// @Failure 500 {object} MessageResponse
// @Router /devices [get]
func (h *Controller) FindAll(c *gin.Context) {
	devices, err := h.service.FindAll()
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
		"message": "Devices found",
		"data":    devices,
	})
}

// FindByID godoc
// @Summary Get device by ID
// @Description Retrieve a single device by its ID.
// @Tags Devices
// @Accept json
// @Produce json
// @Param id path int true "Device ID"
// @Success 200 {object} DetailResponseEnvelope
// @Failure 404 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /devices/{id} [get]
func (h *Controller) FindByID(c *gin.Context) {
	device, err := h.service.FindByID(c.Param("id"))
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
		"message": "Device found",
		"data":    device,
	})
}

// Update godoc
// @Summary Update device
// @Description Update an existing device by its ID.
// @Tags Devices
// @Accept json
// @Produce json
// @Param id path int true "Device ID"
// @Param request body UpdateRequest true "Device payload"
// @Success 200 {object} DetailResponseEnvelope
// @Failure 400 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Failure 409 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /devices/{id} [put]
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
		"message": "Device updated",
		"data":    result,
	})
}

// Delete godoc
// @Summary Delete device
// @Description Delete a device by its ID.
// @Tags Devices
// @Accept json
// @Produce json
// @Param id path int true "Device ID"
// @Success 200 {object} MessageResponse
// @Failure 404 {object} MessageResponse
// @Failure 500 {object} MessageResponse
// @Router /devices/{id} [delete]
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
		"message": "Device deleted",
		"data":    nil,
	})
}

func mapError(err error) int {
	switch {
	case errors.Is(err, ErrDeviceNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrDuplicateDeviceCode):
		return http.StatusConflict
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInvalidStatus):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
