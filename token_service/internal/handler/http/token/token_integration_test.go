package handler_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/1abobik1/token_auth/config"
	"github.com/1abobik1/token_auth/internal/dto"
	handler "github.com/1abobik1/token_auth/internal/handler/http/token"
	service "github.com/1abobik1/token_auth/internal/service/token"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	_ "github.com/lib/pq"
)

type mockPostgesStorage struct {
	tokens map[string]dto.RefreshTokenRecord
}

func newMockPostgesStorage() *mockPostgesStorage {
	return &mockPostgesStorage{
		tokens: make(map[string]dto.RefreshTokenRecord),
	}
}

func (m *mockPostgesStorage) StoreRefreshToken(ctx context.Context, rec dto.RefreshTokenRecord) error {
	key := rec.UserID.String() + rec.JTI
	m.tokens[key] = rec
	return nil
}

func (m *mockPostgesStorage) DeleteRefreshToken(ctx context.Context, userID uuid.UUID, jti string) error {
	key := userID.String() + jti
	delete(m.tokens, key)
	return nil
}

func (m *mockPostgesStorage) GetRefreshToken(ctx context.Context, userID uuid.UUID, jti string) (dto.RefreshTokenRecord, error) {
	key := userID.String() + jti
	token, exists := m.tokens[key]
	if !exists {
		return dto.RefreshTokenRecord{}, sql.ErrNoRows
	}
	if token.ExpiresAt.Before(time.Now()) {
		return dto.RefreshTokenRecord{}, sql.ErrNoRows
	}
	return token, nil
}

func setupRouter(t *testing.T) (*gin.Engine, uuid.UUID) {
	t.Helper()

	cfg := config.Config{
		AccessTTL:  time.Minute,
		RefreshTTL: time.Hour,
		HMACSecret: []byte("secret"),
		CookieTTL:  time.Hour,
	}

	repo := newMockPostgesStorage() 
	svc := service.NewTokenService(repo, cfg)
	h := handler.NewTokenHandler(svc, cfg)

	r := gin.Default()
	r.GET("/token", h.GetTokens)
	r.POST("/token/update", h.Refresh)

	testUserID := uuid.New()
	return r, testUserID
}

func TestFullTokenFlow(t *testing.T) {
	router, userID := setupRouter(t)

	// Step 1: GetTokens
	req := httptest.NewRequest(http.MethodGet, "/token?id="+url.QueryEscape(userID.String()), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	access := extractJSONField(w.Body.String(), "access_token")
	cookies := w.Result().Cookies()
	var refresh string
	for _, c := range cookies {
		if c.Name == "refresh_token" {
			refresh = c.Value
			break
		}
	}
	require.NotEmpty(t, access)
	require.NotEmpty(t, refresh)

	// Step 2: Refresh
	req2 := httptest.NewRequest(http.MethodPost, "/token/update", nil)
	req2.Header.Set("Authorization", "Bearer "+access)
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusOK, w2.Code)
}

func extractJSONField(jsonStr, field string) string {
	start := strings.Index(jsonStr, `"`+field+`":"`)
	if start == -1 {
		return ""
	}
	start += len(field) + 4
	end := strings.Index(jsonStr[start:], `"`)
	if end == -1 {
		return ""
	}
	return jsonStr[start : start+end]
}
