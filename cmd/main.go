package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	handlers "github.com/pabbloacevedog/whatspp-service-glidpa/internal/handlers/http"
	"github.com/pabbloacevedog/whatspp-service-glidpa/internal/usecases"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/config"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/logger"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/utils"
	"github.com/pabbloacevedog/whatspp-service-glidpa/pkg/whatsapp"
	"go.uber.org/zap"
)

func main() {
	// Inicializar el logger
	log, err := logger.New(nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	// No hay método Sync en la interfaz Logger, así que no podemos usar defer log.Sync()

	// Cargar variables de entorno (ignorar error si no existe)
	_ = godotenv.Load() // No falla si el archivo .env no existe

	// Inicializar la configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}
	if cfg.Port == "" {
		log.Fatal("Port not configured")
	}

	// Inicializar el cliente de WhatsApp
	whatsappClient, err := whatsapp.NewClient("./whatsapp.db", whatsapp.WithLogger(log))
	if err != nil {
		log.Fatal("Failed to initialize WhatsApp client", zap.Error(err))
	}

	// Conectar el cliente de WhatsApp
	if err := whatsappClient.Connect(); err != nil {
		log.Fatal("Failed to connect WhatsApp client", zap.Error(err))
	}

	// Inicializar el caso de uso de autenticación
	authUseCase := usecases.NewWhatsAppAuthUseCase(
		whatsappClient,
		log,
		usecases.WithQRTimeout(5*time.Minute),
		usecases.WithQRSize(256),
	)

	// Inicializar el caso de uso de reservas
	bookingUseCase := usecases.NewBookingUseCase(whatsappClient, log)

	// Configurar el router Gin
	router := gin.Default()

	// Configurar CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(cfg.CorsAllowedOrigins, ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Agregar endpoint de health check
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Registrar los manejadores HTTP
	authHandler := handlers.NewAuthHandler(authUseCase, log)
	authHandler.RegisterRoutes(router)

	// Registrar el manejador de reservas
	bookingHandler := handlers.NewBookingHandler(bookingUseCase, log)
	bookingHandler.RegisterRoutes(router, authHandler)

	// Registrar el manejador de webhook para mensajes entrantes
	webhookHandler := handlers.NewWebhookHandler(bookingUseCase, log)
	webhookHandler.RegisterRoutes(router)

	// Configurar el servidor HTTP con timeouts
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Canal para señales de apagado
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Iniciar el servidor en una goroutine
	go func() {
		portInt, _ := strconv.Atoi(cfg.Port)

		// Verificar si el puerto está disponible
		if !utils.IsPortAvailable(portInt) {
			log.Warn("Port is already in use, trying to find an available port", zap.Int("original_port", portInt))

			// Intentar encontrar un puerto disponible (desde el puerto original + 1 hasta el puerto original + 100)
			newPort, err := utils.FindAvailablePort(portInt+1, portInt+100)
			if err != nil {
				log.Fatal("Failed to find an available port", zap.Error(err))
			}

			// Actualizar el puerto en el servidor
			portInt = newPort
			server.Addr = fmt.Sprintf(":%d", portInt)
			log.Info("Using alternative port", zap.Int("new_port", portInt))
		}

		log.Info("Server starting", zap.Int("port", portInt))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed to start", zap.Error(err))
			if err.Error() == "listen tcp :"+strconv.Itoa(portInt)+": bind: address already in use" {
				log.Error("The port is already in use. Please try using a different port by setting the PORT environment variable")
			}
			os.Exit(1) // Salir con código de error en lugar de usar Fatal para permitir un mensaje más descriptivo
		}
	}()

	// Esperar señal de apagado
	<-shutdown
	log.Info("Server stopping")

	// Crear contexto con timeout para el apagado graceful
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Intentar apagar el servidor gracefully
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	// Desconectar el cliente de WhatsApp
	if err := whatsappClient.Disconnect(); err != nil {
		log.Error("Failed to disconnect WhatsApp client", zap.Error(err))
	}

	log.Info("Server stopped")
}
