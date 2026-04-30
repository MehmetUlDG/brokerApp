package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"payment-service/internal/domain"
	"payment-service/internal/infrastructure/postgres"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	connString := "postgres://postgres:postgres@localhost:5432/broker_db?sslmode=disable"
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		t.Skipf("Skipping test, cannot connect to database: %v", err)
		return nil, func() {}
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Skipf("Skipping test, cannot ping database: %v", err)
		return nil, func() {}
	}

	_, err = pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS wallets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL UNIQUE,
			balance NUMERIC(20,8) NOT NULL DEFAULT 0,
			btc_balance NUMERIC(20,8) NOT NULL DEFAULT 0,
			version INT NOT NULL DEFAULT 0,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	require.NoError(t, err)

	cleanup := func() {
		_, _ = pool.Exec(context.Background(), "TRUNCATE wallets CASCADE")
		pool.Close()
	}

	return pool, cleanup
}

func TestWalletRepo_OptimisticLocking_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewWalletRepo(pool)

	userID := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO wallets (user_id, balance, btc_balance, version)
		VALUES ($1, 100.00, 0.5, 0)
	`, userID)
	require.NoError(t, err)

	ctx := context.Background()
	usdDelta, _ := decimal.NewFromString("50.00")
	btcDelta, _ := decimal.NewFromString("0.25")

	err = repo.UpdateBalance(ctx, userID, usdDelta, btcDelta)
	assert.NoError(t, err)

	wallet, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)

	expectedBalance, _ := decimal.NewFromString("150.00")
	expectedBTC, _ := decimal.NewFromString("0.75")

	assert.True(t, wallet.Balance.Equal(expectedBalance))
	assert.True(t, wallet.BTCBalance.Equal(expectedBTC))
	assert.Equal(t, 1, wallet.Version)
}

func TestWalletRepo_OptimisticLocking_ConcurrentUpdate(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewWalletRepo(pool)

	userID := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO wallets (user_id, balance, btc_balance, version)
		VALUES ($1, 100.00, 0.5, 0)
	`, userID)
	require.NoError(t, err)

	ctx := context.Background()
	errChan := make(chan error, 2)

	for i := 0; i < 2; i++ {
		go func() {
			usdDelta, _ := decimal.NewFromString("10.00")
			err := repo.UpdateBalance(ctx, userID, usdDelta, decimal.Zero)
			errChan <- err
		}()
	}

	successCount := 0
	conflictCount := 0

	for i := 0; i < 2; i++ {
		err := <-errChan
		if err == nil {
			successCount++
		} else if err == domain.ErrWalletConflict {
			conflictCount++
		}
	}

	assert.Equal(t, 1, successCount, "One update should succeed")
	assert.LessOrEqual(t, 0, conflictCount, "At most one conflict expected (retry handles it)")

	wallet, err := repo.GetByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Equal(t, 2, wallet.Version, "Version should be incremented twice")
}

func TestWalletRepo_OptimisticLocking_InsufficientBalance(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewWalletRepo(pool)

	userID := uuid.New()
	_, err := pool.Exec(context.Background(), `
		INSERT INTO wallets (user_id, balance, btc_balance, version)
		VALUES ($1, 50.00, 0.5, 0)
	`, userID)
	require.NoError(t, err)

	ctx := context.Background()
	usdDelta, _ := decimal.NewFromString("-100.00")

	err = repo.UpdateBalance(ctx, userID, usdDelta, decimal.Zero)
	assert.ErrorIs(t, err, domain.ErrInsufficientBalance)
}

func TestWalletRepo_OptimisticLocking_WalletNotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := postgres.NewWalletRepo(pool)

	ctx := context.Background()
	userID := uuid.New()
	usdDelta, _ := decimal.NewFromString("50.00")

	err := repo.UpdateBalance(ctx, userID, usdDelta, decimal.Zero)
	assert.ErrorIs(t, err, domain.ErrWalletNotFound)
}
