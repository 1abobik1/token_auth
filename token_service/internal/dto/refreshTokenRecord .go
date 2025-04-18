package dto

import (
	"time"

	"github.com/google/uuid"
)

type RefreshTokenRecord struct {
	UserID    uuid.UUID
	JTI       string
	TokenHash string
	ClientIP  string
	ExpiresAt time.Time
}
