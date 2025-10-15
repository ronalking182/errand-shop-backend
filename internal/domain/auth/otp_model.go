package auth

import (
	"time"

	"gorm.io/gorm"
)

type OneTimePassword struct {
	gorm.Model
	UserID    uint
	Code      string
	ExpiresAt time.Time
}
