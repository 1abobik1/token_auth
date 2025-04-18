package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type AccessClaims struct {
	UserID uuid.UUID `json:"user_id"`
	IP     string    `json:"ip"`
	JTI    string    `json:"jti"`
	jwt.RegisteredClaims
}

// JWT подписанный HMAC‑SHA512
func CreateAccessTokenHS512(userID uuid.UUID, clientIP string, jti string, ttl time.Duration, secret []byte) (string, error) {
	
	now := time.Now()
	claims := AccessClaims{
		UserID: userID,
		IP:     clientIP,
		JTI:    jti,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return tok.SignedString(secret)
}
