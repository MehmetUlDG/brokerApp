package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	pbpayment "payment-service/gen/payment"
	grpcserver "payment-service/internal/grpc"
	httphandler "payment-service/internal/http"
	kafkainfra "payment-service/internal/infrastructure/kafka"
	pginfra "payment-service/internal/infrastructure/postgres"
	stripeinfra "payment-service/internal/infrastructure/stripe"
	"payment-service/internal/usecase"
)

// =============================================================================
// Config
// =============================================================================

type config struct {
	DatabaseURL         string
	KafkaBrokers        []string
	StripeSecretKey     string
	StripeWebhookSecret string
	GRPCPort            string
	HTTPPort            string
}

func loadConfig() config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "50051"
	}

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "8081"
	}

	return config{
		DatabaseURL:         mustEnv("DATABASE_URL"),
		KafkaBrokers:        strings.Split(brokers, ","),
		StripeSecretKey:     mustEnv("STRIPE_SECRET_KEY"),
		StripeWebhookSecret: mustEnv("STRIPE_WEBHOOK_SECRET"),
		GRPCPort:            grpcPort,
		HTTPPort:            httpPort,
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
	return v
}

// =============================================================================
// main
// =============================================================================

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("⚠  .env dosyası yüklenemedi — sistem ortam değişkenleri kullanılıyor")
	}

	logger, _ := zap.NewProduction()
	defer func() {
		if err := logger.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "zap logger sync error: %v\n", err)
		}
	}()

	cfg := loadConfig()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("postgres connection failed", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("postgres ping failed", zap.Error(err))
	}
	logger.Info("postgres connection established")

	txRepo := pginfra.NewTransactionRepo(pool)
	walletRepo := pginfra.NewWalletRepo(pool)

	stripeAdapter := stripeinfra.NewStripeAdapter(cfg.StripeSecretKey)

	publisher := kafkainfra.NewPaymentPublisher(cfg.KafkaBrokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			logger.Warn("kafka writer close error", zap.Error(err))
		}
	}()
	logger.Info("kafka publisher ready", zap.Strings("brokers", cfg.KafkaBrokers))

	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher, logger)

	grpcSrv := grpc.NewServer()
	pbpayment.RegisterPaymentServiceServer(grpcSrv, grpcserver.NewPaymentServer(uc, logger))

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		logger.Fatal("gRPC listener start failed", zap.Error(err))
	}

	go func() {
		logger.Info("gRPC server started", zap.String("port", cfg.GRPCPort))
		if err := grpcSrv.Serve(grpcListener); err != nil {
			logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	mux := http.NewServeMux()
	webhookHandler := httphandler.NewWebhookHandler(txRepo, cfg.StripeWebhookSecret, logger)
	webhookHandler.RegisterRoutes(mux)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("HTTP server started", zap.String("port", cfg.HTTPPort))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	logger.Info("shutdown signal received", zap.String("signal", sig.String()))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP shutdown error", zap.Error(err))
	}

	grpcSrv.GracefulStop()

	if err := publisher.Close(); err != nil {
		logger.Error("Kafka publisher close error", zap.Error(err))
	}

	pool.Close()

	logger.Info("payment-service shutdown completed")
}
