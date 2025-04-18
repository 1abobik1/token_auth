package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/1abobik1/token_auth/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ServiceI interface {
	RefreshTokenPair(ctx context.Context, oldAccess, oldRefresh, clientIP string) (string, string, error)
	IssueTokenPair(ctx context.Context, userID uuid.UUID, clientIP string) (string, string, error)
}

type TokenHandler struct {
	svc ServiceI
	cfg config.Config
}

func NewTokenHandler(svc ServiceI, cfg config.Config) *TokenHandler {
	return &TokenHandler{svc: svc, cfg: cfg}
}

// GET /token?id=<GUID>
func (h *TokenHandler) GetTokens(c *gin.Context) {
	idStr := c.Query("id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	clientIP := c.ClientIP()

	access, refresh, err := h.svc.IssueTokenPair(c, userID, clientIP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue tokens"})
		return
	}

	c.SetCookie("refresh_token", refresh, int(h.cfg.CookieTTL), "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token": access,
	})
}

// POST /token/update
func (h *TokenHandler) Refresh(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization header"})
		return
	}
	oldAccess := parts[1]

	oldRefresh, err := c.Cookie("refresh_token")
	if err != nil || oldRefresh == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing refresh token in cookies"})
		return
	}

	clientIP := c.ClientIP()

	newAccess, newRefresh, err := h.svc.RefreshTokenPair(c, oldAccess, oldRefresh, clientIP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("refresh_token", newRefresh, int(h.cfg.CookieTTL), "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"access_token": newAccess,
	})
}
