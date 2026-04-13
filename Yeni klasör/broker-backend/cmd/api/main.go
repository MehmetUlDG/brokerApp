package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"github.com/yourusername/broker-backend/internal/delivery/http/handler"
	httpmiddleware "github.com/yourusername/broker-backend/internal/delivery/http/middleware"
	"github.com/yourusername/broker-backend/internal/infrastructure/db"
	"github.com/yourusername/broker-backend/internal/infrastructure/outbox"
	"github.com/yourusername/broker-backend/internal/infrastructure/postgres"
	"github.com/yourusername/broker-backend/internal/usecase"
)

func main() {
	// .env dosyasını yükle (geliştirme ortamı için)
	// Production'da sistem ortam değişkenleri kullanılır; bu hata göz ardı edilir.
	if err := godotenv.Load(); err != nil {
		log.Println("⚠  .env dosyası yüklenemedi — sistem ortam değişkenleri kullanılıyor")
	}

	// ── PostgreSQL Bağlantısı ────────────────────────────────────────────────
	maxOpen, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdle, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	lifetime, _ := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"))

	database, err := db.NewPostgresDB(db.Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "broker_user"),
		Password:        getEnv("DB_PASSWORD", "broker_secret"),
		DBName:          getEnv("DB_NAME", "broker_db"),
		SSLMode:         getEnv("DB_SSL_MODE", "disable"),
		MaxOpenConns:    maxOpen,
		MaxIdleConns:    maxIdle,
		ConnMaxLifetime: lifetime,
	})
	if err != nil {
		log.Fatalf("❌ PostgreSQL bağlantı hatası: %v", err)
	}
	defer database.Close()
	log.Println("✅ PostgreSQL bağlantısı kuruldu")

	// ── Repository Katmanı ───────────────────────────────────────────────────
	userRepo   := postgres.NewUserRepository(database)
	walletRepo := postgres.NewWalletRepository(database)
	orderRepo  := postgres.NewOrderRepository(database)

	// ── Usecase Katmanı ──────────────────────────────────────────────────────
	jwtSecret := getEnv("JWT_SECRET", "super_secret_change_this_in_production")
	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		jwtExpiry = 24 * time.Hour
	}

	authUC   := usecase.NewAuthUsecase(userRepo, walletRepo, jwtSecret, jwtExpiry)
	walletUC := usecase.NewWalletUsecase(walletRepo)
	orderUC  := usecase.NewOrderUsecase(orderRepo)

	// ── HTTP Handler Katmanı ─────────────────────────────────────────────────
	authHandler   := handler.NewAuthHandler(authUC)
	walletHandler := handler.NewWalletHandler(walletUC)
	orderHandler  := handler.NewOrderHandler(orderUC)

	// ── HTTP Router (chi) ────────────────────────────────────────────────────
	r := chi.NewRouter()

	// Global middleware stack
	r.Use(chimiddleware.RequestID)  // Her isteğe benzersiz ID
	r.Use(chimiddleware.RealIP)     // X-Forwarded-For / X-Real-IP kullan
	r.Use(chimiddleware.Logger)     // Yapılandırılabilir istek günlüğü
	r.Use(chimiddleware.Recoverer)  // Panic → 500 (servis çökmez)
	r.Use(chimiddleware.Timeout(30 * time.Second)) // Global istek zaman aşımı

	// Health check (herkese açık)
	r.Get("/health", healthCheckHandler(database))

	// Public rotalar (JWT gerekmez)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
	})

	// Korumalı rotalar (JWT zorunlu)
	r.Route("/api", func(r chi.Router) {
		r.Use(httpmiddleware.JWTAuth(jwtSecret))

		// Cüzdan endpoint'leri
		r.Get("/wallet", walletHandler.GetWallet)
		r.Post("/wallet/deposit", walletHandler.Deposit)
		r.Post("/wallet/withdraw", walletHandler.Withdraw)

		// Emir endpoint'leri
		r.Post("/orders", orderHandler.PlaceOrder)
	})

	// ── Context + Graceful Shutdown ──────────────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── Outbox Processor ─────────────────────────────────────────────────────
	kafkaBrokers := getEnv("KAFKA_BROKERS", "localhost:9092")
	ordersTopic  := getEnv("KAFKA_TOPIC_ORDERS", "broker.orders")

	outboxProcessor := outbox.NewProcessor(database, []string{kafkaBrokers}, ordersTopic)
	go outboxProcessor.Start(ctx)
	log.Println("✅ Outbox Processor başlatıldı")

	// ── HTTP Server ──────────────────────────────────────────────────────────
	port := getEnv("HTTP_PORT", "3000")
	readTimeout, _  := time.ParseDuration(getEnv("HTTP_READ_TIMEOUT", "15s"))
	writeTimeout, _ := time.ParseDuration(getEnv("HTTP_WRITE_TIMEOUT", "15s"))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Broker Backend API çalışıyor → http://localhost:%s", port)
		log.Printf("📋 Rotalar: POST /api/auth/register | POST /api/auth/login")
		log.Printf("           GET /api/wallet | POST /api/wallet/deposit | POST /api/wallet/withdraw")
		log.Printf("           POST /api/orders | GET /health")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ HTTP sunucu hatası: %v", err)
		}
	}()

	// OS sinyallerini bekle (CTRL+C veya SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏳ Kapatma sinyali alındı, sunucu düzgünce durduruluyor...")
	cancel() // Outbox Processor ve diğer goroutine'leri durdur

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Graceful shutdown hatası: %v", err)
	}
	log.Println("✅ Sunucu başarıyla kapatıldı")
}

// healthCheckHandler, veritabanı ping'i içeren health endpoint döner.
func healthCheckHandler(database interface{ Ping() error }) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "ok"
		if err := database.Ping(); err != nil {
			dbStatus = fmt.Sprintf("error: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","db":"%s","version":"1.0.0"}`, dbStatus)
	}
}

// getEnv, ortam değişkenini okur; tanımlı değilse defaultVal döner.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
