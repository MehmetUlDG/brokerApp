package domain

import (
	"context"

	"github.com/google/uuid"
)

// =============================================================================
// Auth Domain — Interfaces, Entities ve Errors
// =============================================================================

// UserRepository, kullanıcı veritabanı işlemlerini soyutlar.
// Bağımlılık Tersine Çevirme (DIP): interface domain katmanında,
// implementasyon infrastructure katmanındadır.
type UserRepository interface {
	// Create, yeni bir kullanıcı kaydı oluşturur.
	// E-posta zaten mevcutsa ErrUserAlreadyExists döner.
	Create(ctx context.Context, email, passwordHash, firstName, lastName string) (*User, error)

	// GetByEmail, e-posta adresine göre kullanıcı arar.
	// Bulunamazsa ErrUserNotFound döner.
	GetByEmail(ctx context.Context, email string) (*User, error)

	// GetByID, UUID'ye göre kullanıcı arar.
	// Bulunamazsa ErrUserNotFound döner.
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
}

// AuthUsecase, kimlik doğrulama iş kurallarını soyutlar.
// Delivery katmanı (HTTP handler'lar) bu interface'e bağımlıdır.
type AuthUsecase interface {
	// Register, yeni kullanıcı kaydı oluşturur ve JWT token döner.
	// Kayıt sırasında kullanıcıya sıfır bakiyeli bir cüzdan da oluşturulur.
	Register(ctx context.Context, params RegisterParams) (*User, string, error)

	// Login, kimlik doğrulama yapar ve JWT token döner.
	// Hatalı kimlik bilgilerinde ErrInvalidCredentials döner (bilgi sızdırmaz).
	Login(ctx context.Context, email, password string) (*User, string, error)
}

// RegisterParams, kullanıcı kayıt isteği parametrelerini taşır.
type RegisterParams struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
}

// =============================================================================
// Auth Domain Errors
// =============================================================================

var (
	// ErrUserNotFound, kullanıcı bulunamadığında döner.
	ErrUserNotFound = NewDomainError("user_not_found", "kullanıcı bulunamadı")

	// ErrUserAlreadyExists, e-posta adresi zaten kayıtlıysa döner.
	ErrUserAlreadyExists = NewDomainError("user_already_exists", "bu e-posta adresi zaten kayıtlı")

	// ErrInvalidCredentials, e-posta veya şifre hatalıysa döner.
	// Güvenlik için "hangi alan hatalı" bilgisi verilmez.
	ErrInvalidCredentials = NewDomainError("invalid_credentials", "e-posta veya şifre hatalı")
)
