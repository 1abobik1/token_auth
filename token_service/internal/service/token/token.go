package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/1abobik1/token_auth/config"
	"github.com/1abobik1/token_auth/internal/dto"
	"github.com/1abobik1/token_auth/internal/utils"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidToken = errors.New("invalid access token")
	ErrBadRefresh   = errors.New("refresh token mismatch or expired")
)

type StorageI interface {
	StoreRefreshToken(ctx context.Context, rec dto.RefreshTokenRecord) error
	DeleteRefreshToken(ctx context.Context, userID uuid.UUID, jti string) error
	GetRefreshToken(ctx context.Context, userID uuid.UUID, jti string) (dto.RefreshTokenRecord, error)
}

type TokenService struct {
	repo StorageI
	cfg  config.Config
}

func NewTokenService(repo StorageI, cfg config.Config) *TokenService {
	return &TokenService{repo: repo, cfg: cfg}
}

// Access (JWT HS512) + Refresh (random+bcrypt)
func (s *TokenService) IssueTokenPair(ctx context.Context, userID uuid.UUID, clientIP string) (string, string, error) {
	jti := uuid.NewString()
	access, err := utils.CreateAccessTokenHS512(userID, clientIP, jti, s.cfg.AccessTTL, s.cfg.HMACSecret)
	if err != nil {
		return "", "", err
	}

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	refresh := base64.RawURLEncoding.EncodeToString(raw)

	hash, err := bcrypt.GenerateFromPassword([]byte(refresh), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	expiresAt := time.Now().Add(s.cfg.RefreshTTL)
	rec := dto.RefreshTokenRecord{
		UserID:    userID,
		JTI:       jti,
		TokenHash: string(hash),
		ClientIP:  clientIP,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.StoreRefreshToken(ctx, rec); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

// проверка, ротация и выдача новой пары
func (s *TokenService) RefreshTokenPair(ctx context.Context, oldAccess, oldRefresh, clientIP string) (string, string, error) {
	claims := &utils.AccessClaims{}
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	tok, err := parser.ParseWithClaims(oldAccess, claims, func(t *jwt.Token) (interface{}, error) {
		return s.cfg.HMACSecret, nil
	})
	if err != nil || !tok.Valid {
		return "", "", ErrInvalidToken
	}

	rec, err := s.repo.GetRefreshToken(ctx, claims.UserID, claims.JTI)
	if err != nil {
		return "", "", ErrBadRefresh
	}

	if bcrypt.CompareHashAndPassword([]byte(rec.TokenHash), []byte(oldRefresh)) != nil {
		return "", "", ErrBadRefresh
	}

	if rec.ClientIP != clientIP {
		go utils.SendIPChangeWarningEmail(claims.UserID, rec.ClientIP, clientIP)
	}

	_ = s.repo.DeleteRefreshToken(ctx, claims.UserID, claims.JTI)
	return s.IssueTokenPair(ctx, claims.UserID, clientIP)
}
