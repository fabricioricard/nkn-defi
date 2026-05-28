// cmd/server/main.go
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
	"github.com/fabricioricard/nkn-defi/internal/adapters/nknclient"
	"github.com/fabricioricard/nkn-defi/internal/adapters/postgres"
	redisadapter "github.com/fabricioricard/nkn-defi/internal/adapters/redis"
	"github.com/fabricioricard/nkn-defi/internal/app"
	"github.com/fabricioricard/nkn-defi/internal/engine/amm"
	"github.com/fabricioricard/nkn-defi/internal/engine/pricing"
	"github.com/fabricioricard/nkn-defi/internal/engine/reward"
	"github.com/fabricioricard/nkn-defi/internal/engine/risk"
	"github.com/fabricioricard/nkn-defi/internal/engine/treasury"
	"github.com/fabricioricard/nkn-defi/internal/sync"
	"github.com/fabricioricard/nkn-defi/pkg/logger"
)

func main() {
	cfg := config.Load()
	logg := logger.New(cfg.LogLevel)

	// --- Infraestrutura ---
	db, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Cliente Redis (usado para cache e event bus)
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	defer rdb.Close()

	// Cache (implementação Redis)
	cache := redisadapter.NewCache(rdb)

	// Event Bus (Redis Streams)
	eventBus := redisadapter.NewRedisEventBus(rdb)

	// Cliente NKN Mainnet
	nknClient := nknclient.New(cfg.NKNRPCURL)

	// --- Repositórios ---
	poolRepo := postgres.NewPoolRepository(db)
	swapRepo := postgres.NewSwapRepository(db)
	stakeRepo := postgres.NewStakeRepository(db)
	treasuryRepo := postgres.NewTreasuryRepository(db)
	nodeRepo := postgres.NewNodeRepository(db)

	// --- Engines ---
	ammEngine := amm.NewConstantProductAMM()
	riskEngine := risk.NewEngine()
	pricingEngine := pricing.NewOracle(cache, poolRepo, nknClient)
	_ = reward.NewEngine(swapRepo, poolRepo, eventBus)
	treasuryEngine := treasury.NewEngine(treasuryRepo, eventBus)

	// --- Casos de Uso ---
	poolUC := app.NewPoolUsecase(poolRepo, cache, eventBus, ammEngine)
	swapUC := app.NewSwapUsecase(poolRepo, swapRepo, cache, eventBus, ammEngine, riskEngine, pricingEngine, treasuryEngine)
	stakeUC := app.NewStakeUsecase(stakeRepo, eventBus)
	nodeUC := app.NewNodeUsecase(nodeRepo, eventBus)

	// --- Serviço de Sincronização Blockchain ---
	syncer := sync.NewBlockSyncer(nknClient, eventBus, db, cache, poolUC)
	go syncer.Start(context.Background())

	// --- Servidor HTTP ---
	router := http.NewRouter(poolUC, swapUC, stakeUC, nodeUC, logg, cache)
	srv := http.NewServer(":8080", router)

	go func() {
		logg.Info("Starting HTTP server", zap.String("addr", ":8080"))
		if err := srv.ListenAndServe(); err != nil {
			logg.Fatal("server failed", zap.Error(err))
		}
	}()

	// --- Graceful Shutdown ---
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logg.Info("Shutting down...")
	srv.Shutdown(context.Background())
	syncer.Stop()
}
