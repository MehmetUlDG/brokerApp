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
