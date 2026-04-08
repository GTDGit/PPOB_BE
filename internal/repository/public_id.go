package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	sequenceUserMIC           = "seq_user_mic"
	sequencePublicOrder       = "seq_public_order"
	sequencePublicTransaction = "seq_public_transaction"
	sequencePublicDeposit     = "seq_public_deposit"
	sequencePublicRefund      = "seq_public_refund"
	sequencePublicQRIS        = "seq_public_qris"
)

type sequenceExecutor interface {
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
}

func NewUUID() string {
	return uuid.NewString()
}

func DisplayID(publicID *string, internalID string) string {
	if publicID != nil && *publicID != "" {
		return *publicID
	}
	return internalID
}

func GenerateMIC(ctx context.Context, exec sequenceExecutor) (string, error) {
	return nextFormattedSequence(ctx, exec, sequenceUserMIC, "PID", 6)
}

func GeneratePublicOrderID(ctx context.Context, exec sequenceExecutor) (string, error) {
	return nextFormattedSequence(ctx, exec, sequencePublicOrder, "ORD-", 10)
}

func GeneratePublicTransactionID(ctx context.Context, exec sequenceExecutor) (string, error) {
	return nextFormattedSequence(ctx, exec, sequencePublicTransaction, "IN-", 10)
}

func GeneratePublicDepositID(ctx context.Context, exec sequenceExecutor) (string, error) {
	return nextFormattedSequence(ctx, exec, sequencePublicDeposit, "DP-", 10)
}

func GeneratePublicRefundID(ctx context.Context, exec sequenceExecutor) (string, error) {
	return nextFormattedSequence(ctx, exec, sequencePublicRefund, "RF-", 10)
}

func GeneratePublicQRISID(ctx context.Context, exec sequenceExecutor) (string, error) {
	return nextFormattedSequence(ctx, exec, sequencePublicQRIS, "QR-", 10)
}

func nextFormattedSequence(ctx context.Context, exec sequenceExecutor, sequenceName, prefix string, minWidth int) (string, error) {
	value, err := nextSequenceValue(ctx, exec, sequenceName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%0*d", prefix, minWidth, value), nil
}

func nextSequenceValue(ctx context.Context, exec sequenceExecutor, sequenceName string) (int64, error) {
	var value int64
	row := exec.QueryRowxContext(ctx, `SELECT nextval($1::regclass)`, sequenceName)
	if err := row.Scan(&value); err != nil {
		return 0, err
	}
	return value, nil
}
