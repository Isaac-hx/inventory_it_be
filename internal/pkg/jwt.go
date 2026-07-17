package pkg

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// type object for hold information about  jwt configuration
type JwtConfig struct {
	secretKey string
	expiresIn time.Duration
}

// type object for hold information about claims
type Claims struct {
	UserID         string `json:"user_id"`
	Role           string `json:"role"`
	Email          string `json:"email"`
	Username       string `json:"username"`
	DepartmentName string `json:"department_name"`
	jwt.RegisteredClaims
}

// Constructor for creat jwt config
func NewJwt(secretKey string) *JwtConfig {
	return &JwtConfig{
		secretKey: secretKey,
		expiresIn: 24 * time.Hour,
	}
}

// Method for generate token using jwt config and information in claims object
func (c *JwtConfig) GenerateToken(userID, role, email, username, departmentName string) (string, error) {
	claims := Claims{
		UserID:         userID,
		Role:           role,
		Username:       username,
		Email:          email,
		DepartmentName: departmentName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(c.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(c.secretKey))
}

//Method for parse token and return claims object if token is valid

func (c *JwtConfig) ParseToken(tokenString string) (*Claims, error) {

	///The core proces for parsing tokens;this method conssists of three main steps :
	// 1. parsing
	// 2. Validating & signing
	// 3. Returning the claims
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpectedd signing method %v", t.Header["alg"])
		}
		return []byte(c.secretKey), nil

	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
