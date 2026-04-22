package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
	DatabaseURL          string
	KafkaBrokers         []string
	StripeSecretKey      string
	StripeWebhookSecret  string
	GRPCPort             string
	HTTPPort             string
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
		log.Fatalf("FATAL: environment variable %s is required", key)
	}
	return v
}

// =============================================================================
// main
// =============================================================================

func main() {
	cfg := loadConfig()

	// ── PostgreSQL pool ──────────────────────────────────────────────────────
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("FATAL: cannot connect to postgres: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("FATAL: postgres ping failed: %v", err)
	}
	log.Println("✅ PostgreSQL bağlantısı kuruldu")

	// ── Repos ────────────────────────────────────────────────────────────────
	txRepo := pginfra.NewTransactionRepo(pool)
	walletRepo := pginfra.NewWalletRepo(pool)

	// ── Stripe adapter ───────────────────────────────────────────────────────
	stripeAdapter := stripeinfra.NewStripeAdapter(cfg.StripeSecretKey)

	// ── Kafka publisher ──────────────────────────────────────────────────────
	publisher := kafkainfra.NewPaymentPublisher(cfg.KafkaBrokers)
	defer func() {
		if err := publisher.Close(); err != nil {
			log.Printf("⚠  Kafka writer kapatılırken hata: %v", err)
		}
	}()
	log.Printf("✅ Kafka publisher hazır (brokers=%v)", cfg.KafkaBrokers)

	// ── Usecase ──────────────────────────────────────────────────────────────
	uc := usecase.NewPaymentUsecase(txRepo, walletRepo, stripeAdapter, publisher)

	// ── gRPC server ──────────────────────────────────────────────────────────
	grpcSrv := grpc.NewServer()
	pbpayment.RegisterPaymentServiceServer(grpcSrv, grpcserver.NewPaymentServer(uc))

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GRPCPort))
	if err != nil {
		log.Fatalf("FATAL: gRPC listener başlatılamadı: %v", err)
	}

	go func() {
		log.Printf("🚀 gRPC server başlatıldı — port %s", cfg.GRPCPort)
		if err := grpcSrv.Serve(grpcListener); err != nil {
			log.Printf("❌ gRPC server hatası: %v", err)
		}
	}()

	// ── HTTP server (Stripe webhook) ─────────────────────────────────────────
	mux := http.NewServeMux()
	webhookHandler := httphandler.NewWebhookHandler(txRepo, cfg.StripeWebhookSecret)
	webhookHandler.RegisterRoutes(mux)

	// Health check endpoint
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
		log.Printf("🚀 HTTP server başlatıldı — port %s", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("❌ HTTP server hatası: %v", err)
		}
	}()

	// ── Graceful shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	sig := <-quit
	log.Printf("⏹  Kapatma sinyali alındı: %s — graceful shutdown başlıyor", sig)

	// HTTP server'ı kapat
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠  HTTP shutdown hatası: %v", err)
	}

	// gRPC server'ı kapat (in-flight request'lerin bitmesi beklenir)
	grpcSrv.GracefulStop()

	log.Println("✅ payment-service kapatıldı")
}
