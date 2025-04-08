package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pabbloacevedog/whatspp-service-glidpa/internal/usecases"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/logger"
	"go.uber.org/zap"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authUseCase *usecases.WhatsAppAuthUseCase
	logger      logger.Logger
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authUseCase *usecases.WhatsAppAuthUseCase, logger logger.Logger) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
		logger:      logger,
	}
}

// RegisterRoutes registers the authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	auth := router.Group("/auth")
	{
		auth.GET("/qr", h.GetQR)
		auth.GET("/status", h.GetStatus)
		auth.POST("/logout", h.Logout)
	}
}

// GetQR returns a QR code for authentication
// @Summary Get QR code for authentication
// @Description Returns an SVG QR code for WhatsApp authentication
// @Tags auth
// @Produce text/html
// @Success 200 {string} string "SVG QR code"
// @Failure 400 {object} map[string]string "Error message"
// @Failure 500 {object} map[string]string "Error message"
// @Router /auth/qr [get]
func (h *AuthHandler) GetQR(c *gin.Context) {
	ctx := context.Background()

	// Generate QR code
	qrCode, err := h.authUseCase.GenerateQR(ctx)
	if err != nil {
		h.logger.Error("Failed to generate QR code", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	// Set content type to text/plain for QR code
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, qrCode)
}

// GetStatus returns the current authentication status
// @Summary Get authentication status
// @Description Returns the current WhatsApp authentication status
// @Tags auth
// @Produce json
// @Success 200 {object} usecases.Status "Authentication status"
// @Router /auth/status [get]
func (h *AuthHandler) GetStatus(c *gin.Context) {
	status := h.authUseCase.GetStatus()
	c.JSON(http.StatusOK, status)
}

// Logout logs out from WhatsApp
// @Summary Logout from WhatsApp
// @Description Logs out from WhatsApp and invalidates the session
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "Success message"
// @Failure 500 {object} map[string]string "Error message"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	err := h.authUseCase.Logout()
	if err != nil {
		h.logger.Error("Failed to logout", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// AuthMiddleware is a middleware that checks if the user is authenticated
func (h *AuthHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user is authenticated
		status := h.authUseCase.GetStatus()
		if status.Status != "connected" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		c.Next()
	}
}
