#!/usr/bin/env bash
# =============================================================================
# setup.sh — Demo Broker Backend Project Setup
# Go modülünü başlatır ve tüm bağımlılıkları yükler
# =============================================================================
set -euo pipefail

MODULE_PATH="github.com/yourusername/broker-backend"
GO_VERSION="1.22"

echo "============================================="
echo "  Demo Broker Backend — Project Setup"
echo "============================================="

# 1. Go versiyonu kontrol
echo ""
echo "▶ [1/5] Go sürümü kontrol ediliyor..."
go version || { echo "❌ Go yüklü değil! https://go.dev adresinden yükleyin."; exit 1; }

# 2. Go modülü başlat
echo ""
echo "▶ [2/5] Go modülü initialize ediliyor (go mod init)..."
go mod init "${MODULE_PATH}"

# 3. Bağımlılıkları yükle
echo ""
echo "▶ [3/5] Tüm bağımlılıklar yükleniyor (go get)..."

# HTTP Router
go get github.com/go-chi/chi/v5@latest

# Database
go get github.com/jmoiron/sqlx@latest
go get github.com/lib/pq@latest                      # PostgreSQL driver
go get github.com/google/uuid@latest

# Redis
go get github.com/redis/go-redis/v9@latest

# Kafka
go get github.com/segmentio/kafka-go@latest

# Logging
go get go.uber.org/zap@latest

# JWT
go get github.com/golang-jwt/jwt/v5@latest

# Config / Env
go get github.com/joho/godotenv@latest

# Crypto (bcrypt for password hashing)
go get golang.org/x/crypto@latest

# 4. go.sum güncelle
echo ""
echo "▶ [4/5] go mod tidy çalıştırılıyor..."
go mod tidy

# 5. Tamamlandı
echo ""
echo "▶ [5/5] .env oluşturuluyor..."
cp .env.example .env
echo "  ⚠  .env dosyası oluşturuldu. JWT_SECRET'ı production'da değiştirmeyi unutmayın!"

echo ""
echo "============================================="
echo "  ✅ Setup tamamlandı!"
echo ""
echo "  Altyapıyı başlatmak için:"
echo "    docker compose up -d"
echo ""
echo "  Uygulamayı çalıştırmak için:"
echo "    go run ./cmd/api/..."
echo "============================================="
