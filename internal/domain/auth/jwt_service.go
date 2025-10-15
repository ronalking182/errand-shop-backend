package auth

import (
	"errandShop/internal/pkg/jwt"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService struct {
	secret []byte
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

func (j *JWTService) GenerateToken(user *User) (string, error) {
	// Get user permissions based on role if not set
	permissions := user.Permissions
	if len(permissions) == 0 {
		permissions = GetUserPermissions(user.Role)
	}

	claims := &jwt.JWTClaims{
		UserID:      user.ID,
		Sub:         user.ID.String(),
		Email:       user.Email,
		Name:        user.Name,
		Role:        user.Role,
		Permissions: permissions, // Use actual permissions instead of empty array
		Iat:         time.Now().Unix(),
		Exp:         time.Now().Add(24 * time.Hour).Unix(),
		RegisteredClaims: jwtlib.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwtlib.NewNumericDate(time.Now()),
			ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := jwtlib.MapClaims{
		"sub":  userID.String(),
		"type": "refresh",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTService) ValidateToken(tokenString string) (*jwt.JWTClaims, error) {
	token, err := jwtlib.ParseWithClaims(tokenString, &jwt.JWTClaims{}, func(token *jwtlib.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*jwt.JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
