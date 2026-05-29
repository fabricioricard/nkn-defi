package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"

	"github.com/fabricioricard/nkn-defi/internal/config"
	"github.com/fabricioricard/nkn-defi/internal/adapters/http"
	"github.com/fabricioricard/nkn-defi/internal/adapters/postgres"
	redisadapter "github.com/fabricioricard/nkn-defi/internal/adapters/redis"
	"github.com/fabricioricard/nkn-defi/internal/app"
	"github.com/fabricioricard/nkn-defi/pkg/logger"
)

//go:embed web/dist/*
var frontendFiles embed.FS

func main() {
	cfg := config.Load()
	logg := logger.New(cfg.LogLevel)

	// Banco de dados
	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := postgres.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Redis (cache + event bus)
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	defer rdb.Close()

	cache := redisadapter.NewCache(rdb)
	eventBus := redisadapter.NewRedisEventBus(rdb)

	// Repositórios
	poolRepo := postgres.NewPoolRepository(db)

	// Casos de uso
	poolUC := app.NewPoolUsecase(poolRepo, cache, eventBus)

	// Preparar o sistema de arquivos do frontend
	frontendFS, err := fs.Sub(frontendFiles, "web/dist")
	if err != nil {
		log.Fatalf("failed to setup frontend filesystem: %v", err)
	}

	// Roteador HTTP (agora também serve o frontend)
	router := http.NewRouter(poolUC, logg, frontendFS)
	srv := http.NewServer(":"+cfg.Port(), router)

	go func() {
		logg.Info("Starting HTTP server", zap.String("addr", ":"+cfg.Port()))
		if err := srv.ListenAndServe(); err != nil {
			logg.Fatal("server failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logg.Info("Shutting down...")
	srv.Shutdown(context.Background())
}