package grpcserver

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"payment-service/gen/payment"
	"payment-service/internal/domain"
	"payment-service/internal/usecase"
)

// =============================================================================
// PaymentServer — gRPC handler
// =============================================================================

// PaymentServer, generated PaymentServiceServer interface'ini implement eder.
// Her RPC, usecase katmanına delege edilir.
type PaymentServer struct {
	payment.UnimplementedPaymentServiceServer
	uc     *usecase.PaymentUsecase
	logger *zap.Logger
}

// NewPaymentServer, yeni bir PaymentServer döner.
func NewPaymentServer(uc *usecase.PaymentUsecase, logger *zap.Logger) *PaymentServer {
	return &PaymentServer{uc: uc, logger: logger}
}

// =============================================================================
// Deposit RPC
// =============================================================================

func (s *PaymentServer) Deposit(
	ctx context.Context,
	req *payment.DepositRequest,
) (*payment.DepositResponse, error) {
	if req.UserId == "" || req.Amount == "" || req.Currency == "" || req.StripePaymentMethodId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, amount, currency and stripe_payment_method_id are required")
	}

	tx, err := s.uc.Deposit(ctx, req.UserId, req.Amount, req.Currency, req.StripePaymentMethodId)
	if err != nil {
		s.logger.Error("Deposit RPC failed",
			zap.String("user_id", req.UserId),
			zap.String("amount", req.Amount),
			zap.Error(err))
		if isValidationError(err) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "deposit failed: %v", err)
	}

	s.logger.Info("Deposit RPC success",
		zap.String("user_id", req.UserId),
		zap.String("tx_id", tx.ID.String()))

	return &payment.DepositResponse{
		TransactionId: tx.ID.String(),
		Status:        string(tx.Status),
	}, nil
}

// =============================================================================
// Withdraw RPC
// =============================================================================

func (s *PaymentServer) Withdraw(
	ctx context.Context,
	req *payment.WithdrawRequest,
) (*payment.WithdrawResponse, error) {
	if req.UserId == "" || req.Amount == "" || req.Currency == "" || req.StripeAccountId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id, amount, currency and stripe_account_id are required")
	}

	tx, err := s.uc.Withdraw(ctx, req.UserId, req.Amount, req.Currency, req.StripeAccountId)
	if err != nil {
		s.logger.Error("Withdraw RPC failed",
			zap.String("user_id", req.UserId),
			zap.String("amount", req.Amount),
			zap.Error(err))
		if errors.Is(err, domain.ErrInsufficientBalance) {
			return nil, status.Error(codes.InvalidArgument, "insufficient balance")
		}
		if isValidationError(err) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "withdraw failed: %v", err)
	}

	s.logger.Info("Withdraw RPC success",
		zap.String("user_id", req.UserId),
		zap.String("tx_id", tx.ID.String()))

	return &payment.WithdrawResponse{
		TransactionId: tx.ID.String(),
		Status:        string(tx.Status),
	}, nil
}

// =============================================================================
// GetHistory RPC
// =============================================================================

func (s *PaymentServer) GetHistory(
	ctx context.Context,
	req *payment.HistoryRequest,
) (*payment.HistoryResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.Offset)
	if offset < 0 {
		offset = 0
	}

	txs, err := s.uc.GetHistory(ctx, req.UserId, limit, offset)
	if err != nil {
		if isValidationError(err) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "get history failed: %v", err)
	}

	pbTxs := make([]*payment.Transaction, 0, len(txs))
	for _, tx := range txs {
		pbTxs = append(pbTxs, domainTxToProto(tx))
	}

	return &payment.HistoryResponse{Transactions: pbTxs}, nil
}

// =============================================================================
// GetBalance RPC
// =============================================================================

func (s *PaymentServer) GetBalance(
	ctx context.Context,
	req *payment.BalanceRequest,
) (*payment.BalanceResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	wallet, err := s.uc.GetBalance(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			return nil, status.Error(codes.InvalidArgument, "wallet not found")
		}
		if isValidationError(err) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "get balance failed: %v", err)
	}

	return &payment.BalanceResponse{
		UsdBalance: wallet.Balance.String(),
		BtcBalance: wallet.BTCBalance.String(),
	}, nil
}

// =============================================================================
// Helpers
// =============================================================================

// domainTxToProto, domain.Transaction'ı proto Transaction'a dönüştürür.
// Decimal → string dönüşümü yalnızca bu sınırda gerçekleşir.
func domainTxToProto(tx *domain.Transaction) *payment.Transaction {
	return &payment.Transaction{
		Id:        tx.ID.String(),
		UserId:    tx.UserID.String(),
		Type:      string(tx.Type),
		Amount:    tx.Amount.String(),
		Currency:  tx.Currency,
		Status:    string(tx.Status),
		StripeRef: tx.StripeRef,
		CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// isValidationError, bir hatanın InvalidArgument hata sınıfına girip girmediğini kontrol eder.
func isValidationError(err error) bool {
	var pe *domain.PaymentError
	if errors.As(err, &pe) {
		switch pe.Code {
		case "invalid_amount", "invalid_user_id", "insufficient_balance", "wallet_not_found":
			return true
		}
	}
	return false
}
