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
	"github.com/fabricioricard/nkn-defi/internal/adapters/ethereum"
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
	"github.com/fabricioricard/nkn-defi/internal/app"
	"github.com/fabricioricard/nkn-defi/internal/sync"
	"github.com/fabricioricard/nkn-defi/pkg/logger"
)

//go:embed web/dist/*
var frontendFiles embed.FS

func main() {
	cfg := config.Load()
	logg := logger.New(cfg.LogLevel)

	// ---- Banco de dados ----
	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := postgres.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// ---- Redis (cache + barramento de eventos) ----
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	defer rdb.Close()

	cache := redisadapter.NewCache(rdb)
	eventBus := redisadapter.NewRedisEventBus(rdb)

	// ---- Repositórios ----
	poolRepo := postgres.NewPoolRepository(db)
	bridgeRepo := postgres.NewBridgeRepository(db) // repositório da Bridge

	// ---- Casos de uso ----
	poolUC := app.NewPoolUsecase(poolRepo, cache, eventBus)

	// ---- Cliente NKN (para workers e API) ----
	nknClient := nknclient.New(cfg.NKNRPCURL)

	// ---- Cliente Ethereum (Base) e carregamento do contrato wNKN ----
	ethClient, err := ethereum.NewClient(cfg.BaseRPCURL)
	if err != nil {
		log.Fatalf("failed to create ethereum client: %v", err)
	}
	if err := ethClient.LoadWNKNContract(cfg.WNKNContractAddress); err != nil {
		log.Fatalf("failed to load wNKN contract: %v", err)
	}

	// ---- Workers da Bridge ----
	depositWorker := sync.NewBridgeDepositWorker(db, nknClient, ethClient, bridgeRepo)
	withdrawWorker := sync.NewBridgeWithdrawWorker(db, nknClient, ethClient, bridgeRepo)

	go depositWorker.Start(context.Background())
	go withdrawWorker.Start(context.Background())

	// ---- Frontend embutido ----
	frontendFS, err := fs.Sub(frontendFiles, "web/dist")
	if err != nil {
		log.Fatalf("failed to setup frontend filesystem: %v", err)
	}

	// ---- Roteador HTTP (agora com bridgeRepo) ----
	router := http.NewRouter(poolUC, logg, frontendFS, bridgeRepo)
	srv := http.NewServer(":"+cfg.Port(), router)

	go func() {
		logg.Info("Starting HTTP server", zap.String("addr", ":"+cfg.Port()))
		if err := srv.ListenAndServe(); err != nil {
			logg.Fatal("server failed", zap.Error(err))
		}
	}()

	// ---- Graceful shutdown ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logg.Info("Shutting down...")
	srv.Shutdown(context.Background())
}