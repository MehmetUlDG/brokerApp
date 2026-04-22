-- =============================================================================
-- Demo Broker — TAM VERİTABANI ŞEMASI (Tek Kaynak / Canonical Schema)
-- PostgreSQL 15 | Clean Architecture Backend
-- =============================================================================
-- NOT: Bu dosya tüm tabloların yetkili tanımıdır.
--      001_init_schema.sql ve 002_wallet_service.sql devre dışı bırakılmıştır.
--      Docker Compose initdb dosyaları alfabetik sırayla çalıştırır;
--      bu dosya (000003_*) önce çalışır ve tek geçerli şemayı oluşturur.
-- =============================================================================

-- UUID eklentisi (idempotent)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- TABLE: users
-- Kullanıcı kimlik bilgileri. Şifre asla düz metin saklanmaz (bcrypt ≥ cost 12).
-- =============================================================================
CREATE TABLE IF NOT EXISTS users (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name    VARCHAR(100) NOT NULL DEFAULT '',
    last_name     VARCHAR(100) NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email      ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

COMMENT ON TABLE  users               IS 'Kullanıcı hesapları';
COMMENT ON COLUMN users.id            IS 'UUID primary key (uuid_generate_v4)';
COMMENT ON COLUMN users.email         IS 'Kullanıcı e-postası — UNIQUE kısıtı';
COMMENT ON COLUMN users.password_hash IS 'bcrypt (cost≥12) ile hashlenmiş şifre. Hiçbir zaman düz metin saklanmaz';

-- =============================================================================
-- TABLE: wallets
-- Her kullanıcının tam olarak BİR cüzdanı vardır (user_id UNIQUE).
-- Bakiye alanları DECIMAL(18,8) → float64 hataları yoktur.
-- Race Condition: SELECT ... FOR UPDATE + Transaction (bkz. wallet_repo.go)
-- =============================================================================
CREATE TABLE IF NOT EXISTS wallets (
    id          UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance     DECIMAL(18,8) NOT NULL DEFAULT 0.00000000,  -- USD bakiye
    btc_balance DECIMAL(18,8) NOT NULL DEFAULT 0.00000000,  -- BTC bakiye (demo)
    updated_at  TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT wallets_balance_non_negative     CHECK (balance     >= 0),
    CONSTRAINT wallets_btc_balance_non_negative CHECK (btc_balance >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_user_id    ON wallets(user_id);
CREATE        INDEX IF NOT EXISTS idx_wallets_updated_at ON wallets(updated_at DESC);

COMMENT ON TABLE  wallets             IS 'Kullanıcı cüzdanları (USD + BTC)';
COMMENT ON COLUMN wallets.balance     IS 'USD bakiyesi — DECIMAL(18,8), kayan nokta hatası yok';
COMMENT ON COLUMN wallets.btc_balance IS 'BTC miktarı — DECIMAL(18,8)';
COMMENT ON COLUMN wallets.updated_at  IS 'Son güncelleme zamanı (UTC). Her UpdateBalance çağrısında güncellenir';

-- =============================================================================
-- TABLE: orders
-- Emir kayıtları (BUY/SELL, MARKET/LIMIT).
-- Status: PENDING → matching engine tarafından COMPLETED/FAILED/CANCELED yapılır.
-- =============================================================================
CREATE TABLE IF NOT EXISTS orders (
    id         UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID          NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol     VARCHAR(20)   NOT NULL,                                    -- Örn: 'BTCUSDT'
    side       VARCHAR(10)   NOT NULL CHECK (side   IN ('BUY',  'SELL')),
    type       VARCHAR(10)   NOT NULL CHECK (type   IN ('MARKET', 'LIMIT')),
    price      DECIMAL(18,8) NOT NULL DEFAULT 0.00000000,                 -- LIMIT hedef fiyat; MARKET=0
    quantity   DECIMAL(18,8) NOT NULL,                                    -- Alım/Satım miktarı
    status     VARCHAR(20)   NOT NULL DEFAULT 'PENDING'
                   CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'CANCELED')),
    created_at TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id    ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status     ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_symbol     ON orders(symbol);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC);

COMMENT ON TABLE  orders         IS 'Emir kayıtları. MARKET ve LIMIT emirler desteklenir';
COMMENT ON COLUMN orders.symbol  IS 'İşlem sembolü (örn: BTCUSDT)';
COMMENT ON COLUMN orders.price   IS 'LIMIT emir için hedef fiyat. MARKET emirde 0';
COMMENT ON COLUMN orders.status  IS 'PENDING: beklemede, COMPLETED: gerçekleşti, FAILED/CANCELED: iptal';

-- =============================================================================
-- TABLE: outbox_events
-- Transactional Outbox Pattern — Kafka'ya gönderilmeyi bekleyen eventler.
-- Outbox Processor (5s polling) PENDING satırları okuyup Kafka'ya yazar.
-- FOR UPDATE SKIP LOCKED: Birden fazla processor instance çakışmadan çalışabilir.
-- =============================================================================
CREATE TABLE IF NOT EXISTS outbox_events (
    id             UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_type VARCHAR(50)   NOT NULL,   -- 'ORDER', 'WALLET' vb.
    aggregate_id   VARCHAR(100)  NOT NULL,   -- İlgili kaydın UUID'si (string)
    event_type     VARCHAR(50)   NOT NULL,   -- 'OrderCreated', 'TradeExecuted' vb.
    payload        JSONB         NOT NULL,   -- Event'in tüm verisi (JSONB)
    status         VARCHAR(20)   NOT NULL DEFAULT 'PENDING'
                       CHECK (status IN ('PENDING', 'PROCESSED', 'FAILED')),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Partial index: sadece PENDING satırlar için — Outbox Processor sorgusunu hızlandırır
CREATE INDEX IF NOT EXISTS idx_outbox_pending     ON outbox_events(created_at ASC) WHERE status = 'PENDING';
CREATE INDEX IF NOT EXISTS idx_outbox_event_type  ON outbox_events(event_type);

COMMENT ON TABLE  outbox_events        IS 'Transactional Outbox: Kafka''ya gönderilecek eventler';
COMMENT ON COLUMN outbox_events.status IS 'PENDING→bekliyor, PROCESSED→Kafka''ya gönderildi, FAILED→hata';
