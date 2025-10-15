package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID      uuid.UUID `json:"userID"`
	Sub         string    `json:"sub"` // User ID as string
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Role        string    `json:"role"`
	Permissions []string  `json:"permissions"`
	Iat         int64     `json:"iat"` // Issued at
	Exp         int64     `json:"exp"` // Expires at
}