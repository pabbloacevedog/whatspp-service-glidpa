package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/cdipaolo/sentiment"
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
	Date         string
	EmployeeName string
	PhoneNumber  string
}

// BookingResponse represents the response data for a booking confirmation
type BookingResponse struct {
	BookingID string
	Message   string
	Status    string
}

// MessageResponse represents the response to an incoming message
type MessageResponse struct {
	PhoneNumber string
	Message     string
	Status      string
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
		"¡Hola %s! 😊\n\n"+
			"Tu cita para el servicio de %s está casi lista.\n"+
			"📍 Ubicación: %s\n"+
			"⏰ Hora: %s\n"+
			"📅 Fecha: %s\n"+
			"👤 Atendido por: %s\n\n"+
			"¿Te gustaría confirmar esta cita?\n"+
			"Por favor, responde 'Sí' para confirmar o 'No' para cancelar.\n"+
			"¡Gracias por elegirnos! 🌟",
		request.UserName,
		request.ServiceName,
		request.LocationName,
		request.StartTime,
		request.Date,
		request.EmployeeName,
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

// ProcessIncomingMessage processes incoming messages from WhatsApp
func (u *BookingUseCase) ProcessIncomingMessage(phoneNumber, messageBody string) (*MessageResponse, error) {
	// Check if the client is connected
	if !u.client.IsConnected() {
		return nil, fmt.Errorf("WhatsApp client is not connected")
	}

	// Log the incoming message
	u.logger.Info("Received message from WhatsApp",
		zap.String("phone_number", phoneNumber),
		zap.String("message", messageBody))

	// Parse the phone number to JID format
	jid := types.NewJID(phoneNumber, types.DefaultUserServer)

	// Check if the message is a response to a booking confirmation
	var responseMessage string
	var status string

	// Normalize the message body for case-insensitive comparison
	normalizedMessage := strings.ToLower(messageBody)

	// Variables para el análisis de sentimiento
	var sentimentScore int
	var sentimentAnalysisAvailable bool

	// Inicializar el modelo de análisis de sentimiento
	model, err := sentiment.Restore()
	if err != nil {
		u.logger.Error("Error al cargar el modelo de sentimiento", zap.Error(err))
		// Continuar con el método tradicional si hay error con el modelo
	} else {
		// Realizar análisis de sentimiento (usando inglés como base)
		analysis := model.SentimentAnalysis(normalizedMessage, sentiment.English)
		sentimentScore = int(analysis.Score)
		sentimentAnalysisAvailable = true

		u.logger.Info("Análisis de sentimiento realizado",
			zap.String("mensaje", normalizedMessage),
			zap.Int("score", sentimentScore))

		// Clasificar basado en el sentimiento
		if sentimentScore > 0 {
			u.logger.Info("Sentimiento positivo: Probable confirmación")
		} else if sentimentScore < 0 {
			u.logger.Info("Sentimiento negativo: Probable cancelación")
		} else {
			u.logger.Info("Sentimiento neutro")
		}
	}

	// Mantener el método tradicional de detección por palabras clave
	switch {
	case strings.Contains(normalizedMessage, "sí") || strings.Contains(normalizedMessage, "si"):
		// User confirmed the booking
		responseMessage = "¡Gracias por confirmar tu cita! Te esperamos en la fecha y hora acordada. 😊"
		status = "confirmed"
		u.logger.Info("Usuario confirmó la reserva",
			zap.String("phone_number", phoneNumber),
			zap.String("status", status))

	case strings.Contains(normalizedMessage, "no"):
		// User rejected the booking
		responseMessage = "Hemos cancelado tu cita. Si deseas reagendarla, por favor contáctanos. ¡Gracias!"
		status = "cancelled"
		u.logger.Info("Usuario canceló la reserva",
			zap.String("phone_number", phoneNumber),
			zap.String("status", status))

	default:
		// Unrecognized response - intentar usar el análisis de sentimiento si está disponible
		if sentimentAnalysisAvailable {
			if sentimentScore > 0 {
				// Sentimiento positivo: tratar como confirmación
				responseMessage = "¡Gracias por confirmar tu cita! Te esperamos en la fecha y hora acordada. 😊"
				status = "confirmed"
				u.logger.Info("Usuario confirmó la reserva (por análisis de sentimiento)",
					zap.String("phone_number", phoneNumber),
					zap.String("status", status),
					zap.Int("sentiment_score", sentimentScore))
			} else if sentimentScore < 0 {
				// Sentimiento negativo: tratar como cancelación
				responseMessage = "Hemos cancelado tu cita. Si deseas reagendarla, por favor contáctanos. ¡Gracias!"
				status = "cancelled"
				u.logger.Info("Usuario canceló la reserva (por análisis de sentimiento)",
					zap.String("phone_number", phoneNumber),
					zap.String("status", status),
					zap.Int("sentiment_score", sentimentScore))
			} else {
				// Sentimiento neutro: respuesta no reconocida
				responseMessage = "No entendimos tu respuesta. Por favor, responde 'Sí' para confirmar o 'No' para cancelar tu cita."
				status = "unknown"
				u.logger.Warn("Usuario envió respuesta no reconocida para la reserva",
					zap.String("phone_number", phoneNumber),
					zap.String("message", messageBody),
					zap.String("status", status))
			}
		} else {
			// Si no se pudo cargar el modelo, usar respuesta por defecto
			responseMessage = "No entendimos tu respuesta. Por favor, responde 'Sí' para confirmar o 'No' para cancelar tu cita."
			status = "unknown"
			u.logger.Warn("Usuario envió respuesta no reconocida para la reserva",
				zap.String("phone_number", phoneNumber),
				zap.String("message", messageBody),
				zap.String("status", status))
		}
	}

	// Send response message back to the user
	message := &waE2E.Message{
		Conversation: proto.String(responseMessage),
	}

	// Log before sending message
	u.logger.Info("Intentando enviar respuesta al usuario",
		zap.String("phone_number", phoneNumber),
		zap.String("message", responseMessage),
		zap.String("status", status))

	// Send the message with context
	ctx := context.Background()
	resp, err := u.client.Send(ctx, jid, message)
	if err != nil {
		u.logger.Error("Failed to send response message", zap.Error(err))
		return nil, fmt.Errorf("failed to send response message: %w", err)
	}

	// Log successful message sending
	u.logger.Info("Respuesta enviada exitosamente",
		zap.String("phone_number", phoneNumber),
		zap.String("message_id", resp.ID),
		zap.String("status", status))

	return &MessageResponse{
		PhoneNumber: phoneNumber,
		Message:     responseMessage,
		Status:      status,
	}, nil
}
