package usecases

import (
	"context"
	"fmt"

	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/logger"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/whatsapp"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

// BookingUseCase handles booking-related operations
type BookingUseCase struct {
	client *whatsapp.Client
	logger logger.Logger
}

// NewBookingUseCase creates a new BookingUseCase
func NewBookingUseCase(client *whatsapp.Client, logger logger.Logger) *BookingUseCase {
	return &BookingUseCase{
		client: client,
		logger: logger,
	}
}

// SendConfirmationMessage sends a confirmation message with interactive buttons
func (u *BookingUseCase) SendConfirmationMessage(phoneNumber string) error {
	// Check if the client is connected
	if !u.client.IsConnected() {
		return fmt.Errorf("WhatsApp client is not connected")
	}

	// Parse the phone number to JID format
	jid := types.NewJID(phoneNumber, types.DefaultUserServer)

	// Create a simple text message
	message := &waE2E.Message{
		Conversation: proto.String("¿Deseas confirmar tu cita?\n\nPor favor, responde 'Sí' para confirmar o 'No' para cancelar."),
	}

	// Send the message with context
	ctx := context.Background()
	_, err := u.client.Send(ctx, jid, message)
	if err != nil {
		u.logger.Error("Failed to send confirmation message", zap.Error(err))
		return fmt.Errorf("failed to send confirmation message: %w", err)
	}

	u.logger.Info("Confirmation message sent successfully", zap.String("phone_number", phoneNumber))
	return nil
}
