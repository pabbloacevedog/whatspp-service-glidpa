package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims representa los datos que se almacenarán en el token JWT
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken genera un nuevo token JWT con el user_id proporcionado
func GenerateToken(userID string) (string, error) {
	// Obtener la clave secreta y el tiempo de expiración de las variables de entorno
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return "", errors.New("JWT_SECRET no está configurado")
	}

	expiresStr := os.Getenv("JWT_EXPIRES")
	if expiresStr == "" {
		expiresStr = "1h" // Valor por defecto si no está configurado
	}

	// Parsear la duración de expiración
	expiresDuration, err := time.ParseDuration(expiresStr)
	if err != nil {
		return "", err
	}

	// Crear los claims con el user_id y la información de expiración
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Crear el token con los claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmar el token con la clave secreta
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken valida un token JWT y devuelve los claims si es válido
func ValidateToken(tokenString string) (*Claims, error) {
	// Obtener la clave secreta de las variables de entorno
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		return nil, errors.New("JWT_SECRET no está configurado")
	}

	// Parsear el token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Verificar que el método de firma sea el esperado
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de firma inesperado")
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token inválido")
	}

	return claims, nil
}
