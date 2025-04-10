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
func (h *BookingHandler) RegisterRoutes(router *gin.Engine, authHandler *AuthHandler) {
	booking := router.Group("/booking")
	{
		booking.POST("/confirm", authHandler.AuthMiddleware(), h.ConfirmBooking)
	}
}

// BookingRequest represents the request body for confirming a booking
type BookingRequest struct {
	BookingID    string `json:"booking_id" binding:"required"`
	ServiceName  string `json:"service_name" binding:"required"`
	UserName     string `json:"user_name" binding:"required"`
	LocationName string `json:"location_name" binding:"required"`
	StartTime    string `json:"start_time" binding:"required"`
	Date         string `json:"date" binding:"required"`
	EmployeeName string `json:"employee_name" binding:"required"`
	PhoneNumber  string `json:"phone_number" binding:"required"`
}

// ConfirmBooking sends a confirmation message with booking details
// @Summary Send booking confirmation message
// @Description Sends a WhatsApp message with booking details to the specified number
// @Tags booking
// @Accept json
// @Produce json
// @Param request body BookingRequest true "Booking confirmation request"
// @Success 200 {object} usecases.BookingResponse "Success response"
// @Failure 400 {object} map[string]string "Error message"
// @Failure 500 {object} map[string]string "Error message"
// @Router /booking/confirm [post]
func (h *BookingHandler) ConfirmBooking(c *gin.Context) {
	var request BookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Send confirmation message with booking details
	response, err := h.bookingUseCase.SendConfirmationMessage(usecases.BookingRequest{
		BookingID:    request.BookingID,
		ServiceName:  request.ServiceName,
		UserName:     request.UserName,
		LocationName: request.LocationName,
		StartTime:    request.StartTime,
		Date:         request.Date, // Use 'Date' instead of 'date'
		EmployeeName: request.EmployeeName,
		PhoneNumber:  request.PhoneNumber,
	})

	if err != nil {
		h.logger.Error("Failed to send confirmation message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send confirmation message"})
		return
	}

	c.JSON(http.StatusOK, response)
}
