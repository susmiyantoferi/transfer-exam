package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransferWalletReq struct {
	FromWallet uuid.UUID       `json:"from_wallet" validate:"required"`
	ToWallet   uuid.UUID       `json:"to_wallet" validate:"required"`
	Amount     decimal.Decimal `json:"amount" validate:"required"`
}
