package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pabbloacevedog/whatspp-service-glidpa/internal/usecases"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/logger"
	"go.uber.org/zap"
)

// WebhookHandler handles incoming webhook requests from WhatsApp
type WebhookHandler struct {
	bookingUseCase *usecases.BookingUseCase
	logger         logger.Logger
}

// NewWebhookHandler creates a new WebhookHandler
func NewWebhookHandler(bookingUseCase *usecases.BookingUseCase, logger logger.Logger) *WebhookHandler {
	return &WebhookHandler{
		bookingUseCase: bookingUseCase,
		logger:         logger,
	}
}

// RegisterRoutes registers the webhook routes
func (h *WebhookHandler) RegisterRoutes(router *gin.Engine) {
	router.POST("/webhook", h.HandleIncomingMessage)
}

// WhatsAppMessage represents the structure of an incoming WhatsApp message
type WhatsAppMessage struct {
	From string `json:"from"`
	Body string `json:"body"`
}

// HandleIncomingMessage processes incoming messages from WhatsApp
func (h *WebhookHandler) HandleIncomingMessage(c *gin.Context) {
	var message WhatsAppMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		h.logger.Error("Invalid message format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message format"})
		return
	}

	// Process the message
	response, err := h.bookingUseCase.ProcessIncomingMessage(message.From, message.Body)
	if err != nil {
		h.logger.Error("Failed to process message", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "received",
		"message": response,
	})
}
