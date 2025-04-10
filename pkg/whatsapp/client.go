package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/logger"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"go.uber.org/zap"
)

// WhatsAppMessage represents a message received from WhatsApp
type WhatsAppMessage struct {
	From string
	Body string
}

// EventHandler is a function that handles WhatsApp events
type EventHandler func(evt interface{})

// Client is a wrapper around the whatsmeow client
type Client struct {
	client        *whatsmeow.Client
	store         *sqlstore.Container
	db            *sql.DB
	deviceStore   *store.Device
	handlers      []EventHandler
	handlersMutex sync.RWMutex
	logger        logger.Logger
	connected     bool
	connectedMu   sync.RWMutex
	qrChan        chan string
	qrMutex       sync.RWMutex
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithLogger sets the logger for the client
func WithLogger(logger logger.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// NewClient creates a new WhatsApp client
func NewClient(dbPath string, options ...ClientOption) (*Client, error) {
	// Open the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create the container
	container := sqlstore.NewWithDB(db, "sqlite3", nil)
	if err := container.Upgrade(); err != nil {
		return nil, fmt.Errorf("failed to upgrade database: %w", err)
	}

	// Get the device store
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	// Create the client
	client := &Client{
		store:       container,
		db:          db,
		deviceStore: deviceStore,
		qrChan:      make(chan string, 1), // Buffered channel to prevent blocking
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	// Set default logger if none provided
	if client.logger == nil {
		// zapLogger, _ := zap.NewDevelopment()
		// Use a development logger with default configuration
		devLogger, _ := logger.New(nil)
		client.logger = devLogger
	}

	// Create the whatsmeow client
	client.client = whatsmeow.NewClient(deviceStore, nil)

	// Register event handler
	client.client.AddEventHandler(client.handleEvent)

	return client, nil
}

// Connect connects to WhatsApp
func (c *Client) Connect() error {
	if c.IsConnected() {
		return nil
	}

	err := c.client.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	c.setConnected(true)
	c.logger.Info("Connected to WhatsApp")
	return nil
}

// Disconnect disconnects from WhatsApp
func (c *Client) Disconnect() error {
	if !c.IsConnected() {
		return nil
	}

	c.client.Disconnect()
	c.setConnected(false)
	c.logger.Info("Disconnected from WhatsApp")
	return nil
}

// Logout logs out from WhatsApp and removes the session
func (c *Client) Logout() error {
	// Check if we have a valid device ID
	if c.deviceStore == nil || c.deviceStore.ID == nil {
		c.logger.Warn("No device session found during logout")
		return nil
	}

	// First disconnect if connected
	if c.IsConnected() {
		// Attempt to logout from WhatsApp
		err := c.client.Logout()
		if err != nil {
			c.logger.Error("Failed to logout from WhatsApp", zap.Error(err))
			return fmt.Errorf("failed to logout: %w", err)
		}
		c.setConnected(false)
	}

	// Remove the device from the store
	c.logger.Info("Removing device session", zap.String("device_id", c.deviceStore.ID.String()))
	err := c.deviceStore.Delete()
	if err != nil {
		c.logger.Error("Failed to delete device", zap.Error(err))
		return fmt.Errorf("failed to delete device: %w", err)
	}

	c.logger.Info("Successfully logged out from WhatsApp")
	return nil
}

// IsLoggedIn returns true if the client is logged in
func (c *Client) IsLoggedIn() bool {
	return c.client.Store.ID != nil
}

// IsConnected returns true if the client is connected
func (c *Client) IsConnected() bool {
	c.connectedMu.RLock()
	defer c.connectedMu.RUnlock()
	return c.connected
}

// setConnected sets the connected status
func (c *Client) setConnected(connected bool) {
	c.connectedMu.Lock()
	defer c.connectedMu.Unlock()
	c.connected = connected
}

// GetQRChannel returns a channel that receives QR codes for login
func (c *Client) GetQRChannel(ctx context.Context) <-chan string {
	return c.qrChan
}

// GetPhoneNumber returns the phone number of the logged in user
func (c *Client) GetPhoneNumber() string {
	if !c.IsLoggedIn() {
		return ""
	}
	return c.client.Store.ID.User
}

// AddEventHandler adds an event handler
func (c *Client) AddEventHandler(handler EventHandler) {
	c.handlersMutex.Lock()
	defer c.handlersMutex.Unlock()
	c.handlers = append(c.handlers, handler)
}

// handleEvent handles WhatsApp events
func (c *Client) handleEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Connected:
		c.setConnected(true)
		c.logger.Info("Connected to WhatsApp")

	case *events.Disconnected:
		c.setConnected(false)
		c.logger.Info("Disconnected from WhatsApp")

		// Try to reconnect after a delay
		go func() {
			time.Sleep(5 * time.Second)
			if err := c.Connect(); err != nil {
				c.logger.Error("Failed to reconnect", zap.Error(err))
			}
		}()

	case *events.QR:
		c.qrMutex.Lock()
		// Log QR code event details for debugging
		c.logger.Info("Received QR code event",
			zap.Int("code_count", len(v.Codes)),
			zap.Int("code_length", len(v.Codes[0])),
		// zap.String("timeout", v.Timeout.String())
		)

		// Clear existing QR code if channel is full
		select {
		case <-c.qrChan: // Drain old QR code if present
			c.logger.Info("Cleared old QR code from channel")
		default:
		}

		// Send new QR code
		select {
		case c.qrChan <- v.Codes[0]:
			c.logger.Info("New QR code sent to channel successfully")
		default:
			// This is critical - if we can't send the QR code, the user won't be able to log in
			c.logger.Error("Failed to send QR code to channel - channel is full")
			// Try to drain and send again
			select {
			case <-c.qrChan: // Force drain
				select {
				case c.qrChan <- v.Codes[0]: // Try again
					c.logger.Info("QR code sent to channel after forced drain")
				default:
					c.logger.Error("Still failed to send QR code after forced drain")
				}
			default:
				c.logger.Error("Channel appears empty but send failed - possible deadlock")
			}
		}
		c.qrMutex.Unlock()

	case *events.LoggedOut:
		c.setConnected(false)
		c.logger.Info("Logged out from WhatsApp")

	case *events.Message:
		// Process incoming message
		c.logger.Info("Received message",
			zap.String("from", v.Info.Sender.User),
			zap.String("chat", v.Info.Chat.User),
			zap.String("message_id", v.Info.ID))

		// Extract message content
		var messageBody string
		if v.Message.GetConversation() != "" {
			messageBody = v.Message.GetConversation()
		} else if v.Message.GetExtendedTextMessage() != nil && v.Message.GetExtendedTextMessage().GetText() != "" {
			messageBody = v.Message.GetExtendedTextMessage().GetText()
		}

		if messageBody != "" {
			c.logger.Info("Message content", zap.String("body", messageBody))

			// Create a webhook message
			webhookMessage := &WhatsAppMessage{
				From: v.Info.Sender.User,
				Body: messageBody,
			}

			// Call all registered handlers with the webhook message
			c.handlersMutex.RLock()
			for _, handler := range c.handlers {
				go handler(webhookMessage)
			}
			c.handlersMutex.RUnlock()
		}
	}

	// Call all registered handlers with the original event
	c.handlersMutex.RLock()
	for _, handler := range c.handlers {
		go handler(evt)
	}
	c.handlersMutex.RUnlock()
}

// Send sends a message to the specified JID
func (c *Client) Send(ctx context.Context, jid types.JID, message *waE2E.Message, extra ...whatsmeow.SendRequestExtra) (whatsmeow.SendResponse, error) {
	if !c.IsConnected() {
		return whatsmeow.SendResponse{}, fmt.Errorf("client is not connected")
	}

	// Send the message using the whatsmeow client
	msgID, err := c.client.SendMessage(ctx, jid, message, extra...)
	if err != nil {
		c.logger.Error("Failed to send message", zap.Error(err))
		return whatsmeow.SendResponse{}, fmt.Errorf("failed to send message: %w", err)
	}

	c.logger.Info("Message sent successfully", zap.String("message_id", msgID.ID))
	return msgID, nil
}

// Close closes the client and database connection
func (c *Client) Close() error {
	if c.IsConnected() {
		c.Disconnect()
	}

	return c.db.Close()
}
