package db

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // PostgreSQL driver — side-effect import
)

// Config, PostgreSQL bağlantı parametrelerini tutar.
// Ortam değişkenlerinden doldurulur; sıfır değerler kabul edilmez.
type Config struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewPostgresDB, verilen konfigürasyonla yeni bir sqlx.DB bağlantı havuzu oluşturur.
//
// Davranış:
//   - sqlx.Open() driver'ı kaydeder; gerçek bağlantı Ping ile doğrulanır.
//   - Bağlantı havuzu (connection pool) konfigürasyon değerlerine göre yapılandırılır.
//   - Hata durumunda açık bağlantılar kapatılır.
func NewPostgresDB(cfg Config) (*sqlx.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres driver açılamadı: %w", err)
	}

	// Bağlantı havuzu ayarları
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Gerçek bağlantı sağlık kontrolü
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("postgres bağlantısı doğrulanamadı: %w", err)
	}

	return db, nil
}
