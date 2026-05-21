package delivery

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

// FindAllLogs godoc
// @Summary List delivery logs
// @Description Retrieve all delivery log records.
// @Tags Deliveries
// @Accept json
// @Produce json
// @Success 200 {object} LogsResponseEnvelope
// @Failure 500 {object} MessageResponse
// @Router /delivery-logs [get]
func (h *Controller) FindAllLogs(c *gin.Context) {
	logs, err := h.service.GetDeliveryLogs()
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
		"message": "Success",
		"data":    logs,
	})
}

// SendPending godoc
// @Summary Send pending deliveries
// @Description Process pending monitoring deliveries immediately.
// @Tags Deliveries
// @Accept json
// @Produce json
// @Success 200 {object} ProcessPendingResponse
// @Failure 500 {object} MessageResponse
// @Router /deliveries/send-pending [post]
func (h *Controller) SendPending(c *gin.Context) {
	total, err := h.service.ProcessPending()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pending monitoring data processed",
		"total":   total,
	})
}
