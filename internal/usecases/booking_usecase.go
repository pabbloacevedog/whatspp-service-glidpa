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

// BookingRequest represents the request data for a booking confirmation
type BookingRequest struct {
	BookingID    string
	ServiceName  string
	UserName     string
	LocationName string
	StartTime    string
	PhoneNumber  string
}

// BookingResponse represents the response data for a booking confirmation
type BookingResponse struct {
	BookingID string
	Message   string
	Status    string
}

// SendConfirmationMessage sends a confirmation message with interactive buttons
func (u *BookingUseCase) SendConfirmationMessage(request BookingRequest) (*BookingResponse, error) {
	// Check if the client is connected
	if !u.client.IsConnected() {
		return nil, fmt.Errorf("WhatsApp client is not connected")
	}

	// Parse the phone number to JID format
	jid := types.NewJID(request.PhoneNumber, types.DefaultUserServer)

	// Create a detailed confirmation message
	messageText := fmt.Sprintf(
		"Detalles de tu cita:\n\n"+
			"📅 Servicio: %s\n"+
			"👤 Cliente: %s\n"+
			"📍 Ubicación: %s\n"+
			"⏰ Hora: %s\n\n"+
			"¿Deseas confirmar esta cita?\n"+
			"Por favor, responde 'Sí' para confirmar o 'No' para cancelar.",
		request.ServiceName,
		request.UserName,
		request.LocationName,
		request.StartTime,
	)

	message := &waE2E.Message{
		Conversation: proto.String(messageText),
	}

	// Send the message with context
	ctx := context.Background()
	_, err := u.client.Send(ctx, jid, message)
	if err != nil {
		u.logger.Error("Failed to send confirmation message", zap.Error(err))
		return nil, fmt.Errorf("failed to send confirmation message: %w", err)
	}

	u.logger.Info("Confirmation message sent successfully",
		zap.String("booking_id", request.BookingID),
		zap.String("phone_number", request.PhoneNumber))

	return &BookingResponse{
		BookingID: request.BookingID,
		Message:   messageText,
		Status:    "sent",
	}, nil
}
