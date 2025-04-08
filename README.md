# Bot WhatsApp con IA para Gestión de Citas

Este proyecto implementa un servicio de bot de WhatsApp con integración de IA (Gemini) para la gestión automatizada de citas, desarrollado en Go siguiendo los principios de Clean Architecture.

## Estructura del Proyecto

```
├── cmd
│   └── main.go           # Punto de entrada de la aplicación
├── internal
│   ├── handlers          # Capa de entrega (HTTP/WhatsApp)
│   │   └── http          # Manejadores de endpoints HTTP
│   └── usecases          # Lógica de negocio
│       ├── booking_usecase.go    # Gestión de citas
│       └── whatsapp_auth.go      # Autenticación de WhatsApp
├── pkg
│   ├── config            # Manejo de .env
│   │   └── config.go     # Configuración de la aplicación
│   ├── logger            # Logger estructurado
│   │   └── logger.go     # Implementación del logger
│   ├── utils             # Utilidades generales
│   │   └── port.go       # Gestión de puertos
│   └── whatsapp          # Cliente de WhatsApp
│       └── client.go     # Implementación del cliente
└── deployments           # Archivos Docker
    ├── docker-compose.whatsapp.yml
    └── docker-compose.yml
```

## Endpoints API

### Autenticación WhatsApp

#### GET /auth/qr
- **Descripción**: Obtiene el código QR para autenticación de WhatsApp
- **Respuesta Exitosa**: Código QR en formato SVG
- **Códigos de Error**:
  - 400: Error en la solicitud
  - 500: Error interno del servidor

#### GET /auth/status
- **Descripción**: Obtiene el estado actual de la autenticación de WhatsApp
- **Respuesta Exitosa**: Estado de autenticación en formato JSON

#### POST /auth/logout
- **Descripción**: Cierra la sesión de WhatsApp
- **Respuesta Exitosa**: Mensaje de confirmación
- **Códigos de Error**:
  - 500: Error al cerrar sesión

### Gestión de Citas

#### POST /booking/confirm
- **Descripción**: Envía un mensaje de confirmación con botones interactivos
- **Parámetros Query**:
  - phone_number: Número de teléfono del destinatario (requerido)
- **Respuesta Exitosa**: Mensaje de confirmación
- **Códigos de Error**:
  - 400: Número de teléfono no proporcionado
  - 500: Error al enviar el mensaje

## Configuración

El proyecto utiliza variables de entorno para su configuración. Copia el archivo `.env.example` a `.env` y ajusta los valores según sea necesario.

## Ejecución con Docker

Para ejecutar el servicio usando Docker:

```bash
docker-compose -f deployments/docker-compose.whatsapp.yml up -d
```

El servicio estará disponible en el puerto 8080 por defecto.