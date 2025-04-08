package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pabbloacevedog/whatspp-service-glidpa/internal/usecases"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/logger"
	"go.uber.org/zap"
)

// BookingHandler handles booking-related endpoints
type BookingHandler struct {
	bookingUseCase *usecases.BookingUseCase
	logger         logger.Logger
}

// NewBookingHandler creates a new BookingHandler
func NewBookingHandler(bookingUseCase *usecases.BookingUseCase, logger logger.Logger) *BookingHandler {
	return &BookingHandler{
		bookingUseCase: bookingUseCase,
		logger:         logger,
	}
}

// RegisterRoutes registers the booking routes
func (h *BookingHandler) RegisterRoutes(router *gin.Engine) {
	booking := router.Group("/booking")
	{
		booking.POST("/confirm", h.ConfirmBooking)
	}
}

// ConfirmBooking sends a confirmation message with interactive buttons
// @Summary Send booking confirmation message
// @Description Sends a WhatsApp message with confirmation buttons to the specified number
// @Tags booking
// @Accept json
// @Produce json
// @Param phone_number query string true "Phone number to send confirmation to"
// @Success 200 {object} map[string]string "Success message"
// @Failure 400 {object} map[string]string "Error message"
// @Failure 500 {object} map[string]string "Error message"
// @Router /booking/confirm [post]
func (h *BookingHandler) ConfirmBooking(c *gin.Context) {
	phoneNumber := c.Query("phone_number")
	if phoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is required"})
		return
	}

	// Send confirmation message with buttons
	err := h.bookingUseCase.SendConfirmationMessage(phoneNumber)
	if err != nil {
		h.logger.Error("Failed to send confirmation message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send confirmation message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Confirmation message sent successfully"})
}
