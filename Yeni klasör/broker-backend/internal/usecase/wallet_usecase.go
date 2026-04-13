package usecase

import (
	"context"
	"fmt"

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
//   - BUY  → USD = quantity×price düşer, BTC = quantity artar
//   - SELL → BTC = quantity düşer, USD = quantity×price artar
//
// NOT: İki ayrı UpdateBalance çağrısı yapılır; her biri kendi transaction'ını açar.
// Production'da WalletRepository'ye atomik tek-tx metodu eklenebilir.
func (u *walletUsecase) TransferForOrder(
	ctx context.Context,
	userID uuid.UUID,
	side string,
	quantity, price decimal.Decimal,
) (*domain.Wallet, error) {
	total := quantity.Mul(price)

	switch side {
	case "BUY":
		// USD düşer
		if _, err := u.repo.UpdateBalance(ctx, domain.UpdateBalanceParams{
			UserID: userID,
			Field:  domain.BalanceFieldUSD,
			Amount: total,
			Type:   domain.TransferTypeDebit,
		}); err != nil {
			return nil, fmt.Errorf("BUY - USD düşülemedi: %w", err)
		}
		// BTC artar
		return u.repo.UpdateBalance(ctx, domain.UpdateBalanceParams{
			UserID: userID,
			Field:  domain.BalanceFieldBTC,
			Amount: quantity,
			Type:   domain.TransferTypeCredit,
		})

	case "SELL":
		// BTC düşer
		if _, err := u.repo.UpdateBalance(ctx, domain.UpdateBalanceParams{
			UserID: userID,
			Field:  domain.BalanceFieldBTC,
			Amount: quantity,
			Type:   domain.TransferTypeDebit,
		}); err != nil {
			return nil, fmt.Errorf("SELL - BTC düşülemedi: %w", err)
		}
		// USD artar
		return u.repo.UpdateBalance(ctx, domain.UpdateBalanceParams{
			UserID: userID,
			Field:  domain.BalanceFieldUSD,
			Amount: total,
			Type:   domain.TransferTypeCredit,
		})

	default:
		return nil, fmt.Errorf("geçersiz emir yönü: %s", side)
	}
}
