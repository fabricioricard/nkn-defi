package main

import (
    "context"
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

func main() {
    cfg := config.Load()
    logg := logger.New(cfg.LogLevel)

    db, err := postgres.Connect(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }
    defer db.Close()

    // Executa migrações
    if err := postgres.RunMigrations(db); err != nil {
        log.Fatalf("failed to run migrations: %v", err)
    }

    rdb := redis.NewClient(&redis.Options{
        Addr: cfg.RedisURL,
    })
    defer rdb.Close()

    cache := redisadapter.NewCache(rdb)
    eventBus := redisadapter.NewRedisEventBus(rdb)

    // Repos
    poolRepo := postgres.NewPoolRepository(db)

    // Usecases (stubs para os outros ainda não implementados)
    poolUC := app.NewPoolUsecase(poolRepo, cache, eventBus)

    // Servidor HTTP
    router := http.NewRouter(poolUC, logg)
    srv := http.NewServer(":"+cfg.Port(), router)

    go func() {
        logg.Info("Starting HTTP server", zap.String("addr", ":"+cfg.Port()))
        if err := srv.ListenAndServe(); err != nil {
            logg.Fatal("server failed", zap.Error(err))
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    logg.Info("Shutting down...")
    srv.Shutdown(context.Background())
}