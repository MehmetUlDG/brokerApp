package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourusername/broker-backend/internal/domain"
)

// authUsecase, domain.AuthUsecase arayüzünün implementasyonudur.
type authUsecase struct {
	userRepo   domain.UserRepository
	walletRepo domain.WalletRepository
	jwtSecret  string
	jwtExpiry  time.Duration
}

// NewAuthUsecase, yeni bir authUsecase örneği döner.
//
// Parametreler:
//   - userRepo   : Kullanıcı repository'si
//   - walletRepo : Kayıt sırasında cüzdan oluşturmak için
//   - jwtSecret  : JWT imzalama anahtarı (üretimde güçlü rassal değer olmalı)
//   - jwtExpiry  : Token geçerlilik süresi (örn. 24h)
func NewAuthUsecase(
	userRepo domain.UserRepository,
	walletRepo domain.WalletRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
) domain.AuthUsecase {
	return &authUsecase{
		userRepo:   userRepo,
		walletRepo: walletRepo,
		jwtSecret:  jwtSecret,
		jwtExpiry:  jwtExpiry,
	}
}

// Register, yeni kullanıcı kaydı oluşturur ve kullanıcıya sıfır bakiyeli
// bir cüzdan açar. Başarıda kullanıcı ve JWT token döner.
//
// Güvenlik: Şifre bcrypt (cost=12) ile hashlenir; düz metin asla saklanmaz.
func (a *authUsecase) Register(ctx context.Context, params domain.RegisterParams) (*domain.User, string, error) {
	// 1. Şifreyi bcrypt ile hashle (cost=12 → güvenli, ~250ms/hash)
	hash, err := bcrypt.GenerateFromPassword([]byte(params.Password), 12)
	if err != nil {
		return nil, "", fmt.Errorf("şifre hashlenemedi: %w", err)
	}

	// 2. Kullanıcıyı ve cüzdanını tek bir transaction içinde (atomik) kaydet
	user, err := a.userRepo.CreateUserWithWallet(ctx, params.Email, string(hash), params.FirstName, params.LastName)
	if err != nil {
		return nil, "", err // domain.ErrUserAlreadyExists vb.
	}

	// 4. JWT oluştur ve döndür
	token, err := a.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// Login, e-posta ve şifreyi doğrular; başarıda kullanıcı ve JWT döner.
//
// Güvenlik: Kullanıcı bulunamasa da "invalid_credentials" dönülür.
// Bu sayede saldırgan hangi e-postanın kayıtlı olduğunu öğrenemez.
func (a *authUsecase) Login(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := a.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Kullanıcı bulunamasa bile bilgi sızdırmayı önlemek için ErrInvalidCredentials
		return nil, "", domain.ErrInvalidCredentials
	}

	// Şifre karşılaştırması (timing-safe)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", domain.ErrInvalidCredentials
	}

	token, err := a.generateToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

// generateToken, kullanıcı için HS256 imzalı JWT oluşturur.
//
// Claims:
//   - sub : User UUID (string)
//   - exp : Unix timestamp (expiry)
//   - iat : Unix timestamp (issued at)
func (a *authUsecase) generateToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID.String(),
		"exp": time.Now().Add(a.jwtExpiry).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("JWT imzalanamadı: %w", err)
	}
	return signed, nil
}
