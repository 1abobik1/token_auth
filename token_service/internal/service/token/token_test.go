package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/1abobik1/token_auth/config"
	"github.com/1abobik1/token_auth/internal/dto"
	service "github.com/1abobik1/token_auth/internal/service/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStorage struct {
	storeFunc  func(dto.RefreshTokenRecord) error
	getFunc    func(uuid.UUID, string) (dto.RefreshTokenRecord, error)
	deleteFunc func(uuid.UUID, string) error
}

func (m *mockStorage) StoreRefreshToken(ctx context.Context, rec dto.RefreshTokenRecord) error {
	return m.storeFunc(rec)
}

func (m *mockStorage) GetRefreshToken(ctx context.Context, userID uuid.UUID, jti string) (dto.RefreshTokenRecord, error) {
	return m.getFunc(userID, jti)
}

func (m *mockStorage) DeleteRefreshToken(ctx context.Context, userID uuid.UUID, jti string) error {
	return m.deleteFunc(userID, jti)
}

func TestIssueTokenPair(t *testing.T) {
	mockRepo := &mockStorage{
		storeFunc: func(rec dto.RefreshTokenRecord) error {
			return nil
		},
	}
	cfg := config.Config{
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
		HMACSecret: []byte("secret"),
	}
	svc := service.NewTokenService(mockRepo, cfg)

	userID := uuid.New()
	access, refresh, err := svc.IssueTokenPair(context.Background(), userID, "127.0.0.1")

	require.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}
