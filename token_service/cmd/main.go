package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1abobik1/token_auth/config"
	service "github.com/1abobik1/token_auth/internal/service/token"
	handler "github.com/1abobik1/token_auth/internal/handler/http/token"
	"github.com/1abobik1/token_auth/internal/storage/postgresql"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.MustLoad()

	repo, err := postgresql.NewPostgresStorageProd(cfg.StoragePath)
	if err != nil {
		panic("postgres connection error: " + err.Error())
	}
	defer func() {
		if err := repo.Close(); err != nil {
			panic("failed to close DB: " + err.Error())
		}
	}()

	tokenService := service.NewTokenService(repo, *cfg)
	tokenHandler := handler.NewTokenHandler(tokenService, *cfg)

	router := gin.Default()
	router.GET("/token", tokenHandler.GetTokens)
	router.POST("/token/update", tokenHandler.Refresh)

	server := &http.Server{
		Addr:    cfg.HTTPServer,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("server error: " + err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		panic("server shutdown error: " + err.Error())
	}
}