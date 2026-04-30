package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/yourusername/broker-backend/internal/domain"
)

// walletUsecase, domain.WalletUsecase arayüzünün implementasyonudur.
type walletUsecase struct {
	repo domain.WalletRepository
}

// NewWalletUsecase, yeni bir walletUsecase örneği döner.
func NewWalletUsecase(repo domain.WalletRepository) domain.WalletUsecase {
	return &walletUsecase{repo: repo}
}

// GetWallet, kullanıcının cüzdanını döner.
func (u *walletUsecase) GetWallet(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	return u.repo.GetByUserID(ctx, userID)
}

// Deposit, kullanıcının USD bakiyesine amount ekler.
// amount > 0 olmalıdır, aksi hâlde ErrInvalidAmount döner.
func (u *walletUsecase) Deposit(
	ctx context.Context,
	userID uuid.UUID,
	amount decimal.Decimal,
) (*domain.Wallet, error) {
	if amount.IsNegative() || amount.IsZero() {
		return nil, domain.ErrInvalidAmount
	}
	return u.repo.UpdateBalance(ctx, domain.UpdateBalanceParams{
		UserID: userID,
		Field:  domain.BalanceFieldUSD,
		Amount: amount,
		Type:   domain.TransferTypeCredit,
	})
}

// Withdraw, kullanıcının USD bakiyesinden amount düşer.
// Yetersiz bakiye durumunda ErrInsufficientBalance döner.
func (u *walletUsecase) Withdraw(
	ctx context.Context,
	userID uuid.UUID,
	amount decimal.Decimal,
) (*domain.Wallet, error) {
	if amount.IsNegative() || amount.IsZero() {
		return nil, domain.ErrInvalidAmount
	}
	return u.repo.UpdateBalance(ctx, domain.UpdateBalanceParams{
		UserID: userID,
		Field:  domain.BalanceFieldUSD,
		Amount: amount,
		Type:   domain.TransferTypeDebit,
	})
}

// TransferForOrder, emir gerçekleştirme sırasında çağrılır.
// Tüm işlemler artık tek bir veritabanı transaction'ı içinde atomik olarak gerçekleşir.
func (u *walletUsecase) TransferForOrder(
	ctx context.Context,
	userID uuid.UUID,
	side string,
	quantity, price decimal.Decimal,
) (*domain.Wallet, error) {
	return u.repo.TransferForOrder(ctx, userID, side, quantity, price)
}
