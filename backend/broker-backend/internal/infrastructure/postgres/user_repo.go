package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/yourusername/broker-backend/internal/domain"
)

// userRepository, domain.UserRepository arayüzünün PostgreSQL implementasyonudur.
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository, yeni bir userRepository örneği döner.
// db nil olamaz; aksi hâlde panic oluşur.
func NewUserRepository(db *sqlx.DB) domain.UserRepository {
	if db == nil {
		panic("user repository: sqlx.DB nil olamaz")
	}
	return &userRepository{db: db}
}

// Create, yeni kullanıcı kaydı oluşturur.
// E-posta adresi benzersizliği DB tarafından korunur.
// Duplicate e-posta → domain.ErrUserAlreadyExists
func (r *userRepository) Create(
	ctx context.Context,
	email, passwordHash, firstName, lastName string,
) (*domain.User, error) {
	const query = `
		INSERT INTO users (id, email, password_hash, first_name, last_name, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, email, password_hash, first_name, last_name, created_at
	`

	var u domain.User
	err := r.db.QueryRowxContext(
		ctx, query,
		uuid.New(), email, passwordHash, firstName, lastName,
	).StructScan(&u)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("user Create: %w", err)
	}
	return &u, nil
}

// CreateUserWithWallet, yeni bir kullanıcı kaydı ve ona bağlı sıfır bakiyeli
// cüzdanı tek bir veritabanı transaction'ı içinde atomik olarak oluşturur.
func (r *userRepository) CreateUserWithWallet(
	ctx context.Context,
	email, passwordHash, firstName, lastName string,
) (*domain.User, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("user CreateUserWithWallet: transaction başlatılamadı: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	rollback := func(cause error) error {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("%w; rollback hatası: %v", cause, rbErr)
		}
		return cause
	}

	// 1. Kullanıcıyı oluştur
	const userQuery = `
		INSERT INTO users (id, email, password_hash, first_name, last_name, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, email, password_hash, first_name, last_name, created_at
	`

	newUserID := uuid.New()
	var u domain.User
	err = tx.QueryRowxContext(
		ctx, userQuery,
		newUserID, email, passwordHash, firstName, lastName,
	).StructScan(&u)

	if err != nil {
		if isUniqueViolation(err) {
			return nil, rollback(domain.ErrUserAlreadyExists)
		}
		return nil, rollback(fmt.Errorf("user CreateUserWithWallet (user insert): %w", err))
	}

	// 2. Cüzdanı oluştur
	const walletQuery = `
		INSERT INTO wallets (id, user_id, balance, btc_balance, updated_at)
		VALUES ($1, $2, 0, 0, NOW())
	`

	newWalletID := uuid.New()
	_, err = tx.ExecContext(ctx, walletQuery, newWalletID, newUserID)
	if err != nil {
		return nil, rollback(fmt.Errorf("user CreateUserWithWallet (wallet insert): %w", err))
	}

	// 3. Commit
	if err = tx.Commit(); err != nil {
		return nil, rollback(fmt.Errorf("user CreateUserWithWallet: commit hatası: %w", err))
	}

	return &u, nil
}

// GetByEmail, e-posta adresine göre kullanıcı getirir.
// Bulunamazsa domain.ErrUserNotFound döner.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, email, password_hash, first_name, last_name, created_at
		FROM   users
		WHERE  email = $1
		LIMIT  1
	`

	var u domain.User
	if err := r.db.GetContext(ctx, &u, query, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("user GetByEmail: %w", err)
	}
	return &u, nil
}

// GetByID, UUID'ye göre kullanıcı getirir.
// Bulunamazsa domain.ErrUserNotFound döner.
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const query = `
		SELECT id, email, password_hash, first_name, last_name, created_at
		FROM   users
		WHERE  id = $1
		LIMIT  1
	`

	var u domain.User
	if err := r.db.GetContext(ctx, &u, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("user GetByID: %w", err)
	}
	return &u, nil
}
