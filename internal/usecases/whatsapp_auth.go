package usecases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/logger"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/whatsapp"
	"go.uber.org/zap"
)

// WhatsAppAuthUseCase handles WhatsApp authentication
type WhatsAppAuthUseCase struct {
	client      *whatsapp.Client
	logger      logger.Logger
	qrTimeout   time.Duration
	qrSize      int
	qrCodeCache string
	// QRCodeCache is exported for testing purposes
	QRCodeCache string
}

// WhatsAppAuthUseCaseOption is a function that configures a WhatsAppAuthUseCase
type WhatsAppAuthUseCaseOption func(*WhatsAppAuthUseCase)

// WithQRTimeout sets the QR code timeout
func WithQRTimeout(timeout time.Duration) WhatsAppAuthUseCaseOption {
	return func(u *WhatsAppAuthUseCase) {
		u.qrTimeout = timeout
	}
}

// WithQRSize sets the QR code size
func WithQRSize(size int) WhatsAppAuthUseCaseOption {
	return func(u *WhatsAppAuthUseCase) {
		u.qrSize = size
	}
}

// NewWhatsAppAuthUseCase creates a new WhatsAppAuthUseCase
func NewWhatsAppAuthUseCase(client *whatsapp.Client, logger logger.Logger, options ...WhatsAppAuthUseCaseOption) *WhatsAppAuthUseCase {
	useCase := &WhatsAppAuthUseCase{
		client:    client,
		logger:    logger,
		qrTimeout: 5 * time.Minute,
		qrSize:    256,
	}

	// Apply options
	for _, option := range options {
		option(useCase)
	}

	return useCase
}

// Status represents the authentication status
type Status struct {
	Status string `json:"status"`
	Phone  string `json:"phone,omitempty"`
}

// GetStatus returns the current authentication status
func (u *WhatsAppAuthUseCase) GetStatus() Status {
	if u.client.IsLoggedIn() {
		return Status{
			Status: "connected",
			Phone:  u.client.GetPhoneNumber(),
		}
	}

	return Status{
		Status: "disconnected",
	}
}

// GenerateQR generates a QR code for authentication
func (u *WhatsAppAuthUseCase) GenerateQR(ctx context.Context) (string, error) {
	// If already logged in, return an error
	if u.client.IsLoggedIn() {
		return "", errors.New("already logged in")
	}

	// Clear the QR code cache to ensure we get a fresh QR code
	u.qrCodeCache = ""
	u.QRCodeCache = ""
	u.logger.Info("Generating new QR code for authentication")

	// Connect to WhatsApp if not connected
	if !u.client.IsConnected() {
		u.logger.Info("Connecting to WhatsApp for QR code generation")
		if err := u.client.Connect(); err != nil {
			u.logger.Error("Failed to connect to WhatsApp", zap.Error(err))
			return "", fmt.Errorf("failed to connect to WhatsApp: %w", err)
		}
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, u.qrTimeout)
	defer cancel()

	// Get the QR channel
	qrChan := u.client.GetQRChannel(ctx)
	u.logger.Info("Waiting for QR code from WhatsApp")

	// Wait for a QR code
	select {
	case qrCode := <-qrChan:
		// Validate QR code
		if qrCode == "" {
			u.logger.Error("Received empty QR code from WhatsApp")
			return "", errors.New("received empty QR code from WhatsApp")
		}

		// Cache the QR code text
		u.qrCodeCache = qrCode
		u.QRCodeCache = qrCode
		u.logger.Info("Successfully received and cached QR code",
			zap.Int("qr_code_length", len(qrCode)))

		return qrCode, nil

	case <-ctx.Done():
		u.logger.Error("Timeout waiting for QR code from WhatsApp")
		return "", errors.New("timeout waiting for QR code")
	}
}

// Logout logs out from WhatsApp
func (u *WhatsAppAuthUseCase) Logout() error {
	// Clear the QR code cache
	u.qrCodeCache = ""
	u.QRCodeCache = ""

	// Logout from WhatsApp
	return u.client.Logout()
}
