package config

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	SecretKey   string
	ExpiredHour int
}

// Generate token using
func GenerateJWT(config JWTConfig, username string, email string, role string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"email":    email,
		"role":     role,
		"exp":      time.Now().Add(time.Duration(config.ExpiredHour) * time.Hour).Unix(),
		"iss":      "inventory-it",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// VerifyToken verifies the JWT token and returns the claims if valid
func VerifyToken(tokenString string, secretKey string) (jwt.MapClaims, error) {
	//Check if the token is signing method is HMAC, and return the secret key for validation
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("cannot parse claims")
	}

	// validasi field wajib
	username, ok1 := claims["username"].(string)
	email, ok2 := claims["email"].(string)
	role, ok3 := claims["role"].(string)

	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("invalid claims data")
	}

	// optional: bisa log / debug
	_ = username
	_ = email
	_ = role

	return claims, nil
}
