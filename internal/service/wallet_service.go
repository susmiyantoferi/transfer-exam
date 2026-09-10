package service

import (
	"context"
	"errors"
	"transfer-exam/internal/dto"
	"transfer-exam/internal/entity"
	"transfer-exam/internal/repository"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletService interface {
	TransferWallet(c context.Context, req *dto.TransferWalletReq) (uuid.UUID, error)
}

type walletServiceImpl struct {
	WalletRepo repository.WalletRepository
	Db         *gorm.DB
	Log        *logrus.Logger
	Validate   *validator.Validate
}

func NewWalletServiceImpl(walletRepo repository.WalletRepository, db *gorm.DB, log *logrus.Logger, validate *validator.Validate) WalletService {
	return &walletServiceImpl{
		WalletRepo: walletRepo,
		Db:         db,
		Log:        log,
		Validate:   validate,
	}
}

var (
	ErrAmountMustBeGreater = errors.New("amount must be greater than zero")
	ErrCannotSameAccount   = errors.New("cannot transfer same account")
	ErrInvalidCurrency     = errors.New("transfer currency invalid")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrValidate            = errors.New("validation failed")
)

func (w *walletServiceImpl) TransferWallet(c context.Context, req *dto.TransferWalletReq) (uuid.UUID, error) {
	traceID := uuid.New()

	if err := w.Validate.Struct(req); err != nil {
		w.Log.WithField("trace_id", traceID).WithError(err).Error("transfer: validation failed")
		return traceID, err
	}

	//rounded amount
	amount := req.Amount.Round(2)
	if amount.LessThanOrEqual(decimal.Zero) {
		return traceID, ErrAmountMustBeGreater
	}

	//cekk not same wallet tranfer
	if req.FromWallet == req.ToWallet {
		return traceID, ErrCannotSameAccount
	}

	if err := w.Db.WithContext(c).Transaction(func(tx *gorm.DB) error {

		firstID := req.FromWallet
		secondID := req.ToWallet
		refID := xid.New().String()

		//set first lock n secc lock
		if req.FromWallet.String() > req.ToWallet.String() {
			firstID = req.ToWallet
			secondID = req.FromWallet
		}

		//lock wallet
		var firstWallet entity.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&firstWallet, firstID).Error; err != nil {
			w.Log.WithField("trace_id", traceID).WithError(err).Error("transfer: failed get from wallet")
			return gorm.ErrRecordNotFound
		}

		var seccondWallet entity.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seccondWallet, secondID).Error; err != nil {
			w.Log.WithField("trace_id", traceID).WithError(err).Error("transfer: failed get wallet")
			return gorm.ErrRecordNotFound
		}

		//sett receiver n sender
		var fromWallet, toWallet entity.Wallet
		if firstWallet.ID == req.FromWallet {
			fromWallet = firstWallet
			toWallet = seccondWallet
		} else {
			fromWallet = seccondWallet
			toWallet = firstWallet
		}

		//cekk currency
		if fromWallet.Currency != toWallet.Currency {
			w.Log.WithField("trace_id", traceID).Warn("transfer currency invalid")
			return ErrInvalidCurrency
		}

		//cek balance
		if fromWallet.Balance.LessThan(amount) {
			w.Log.WithField("trace_id", traceID).Warn("transfer: insufficient balance")
			return ErrInsufficientBalance
		}

		//cekk for idempotency
		var led entity.Ledger
		err := tx.Where("reference_id = ?", refID).First(&led).Error
		if err == nil {
			w.Log.WithField("trace_id", traceID).Infof("transfer: duplicate request ignored, refID: %s", refID)
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			w.Log.WithField("trace_id", traceID).WithError(err).Error("transfer: failed check idempotency")
			return err
		}

		mins := map[string]any{
			"balance": gorm.Expr("balance - ?", amount),
		}

		plus := map[string]any{
			"balance": gorm.Expr("balance + ?", amount),
		}

		if err := tx.Model(&fromWallet).Updates(mins).Error; err != nil {
			w.Log.WithField("trace_id", traceID).WithError(err).Error("transfer: failed reduce balance")
			return err
		}

		if err := tx.Model(&toWallet).Updates(plus).Error; err != nil {
			w.Log.WithField("trace_id", traceID).WithError(err).Error("transfer: failed plus balance")
			return err
		}

		minsLedger := entity.Ledger{
			WalletID:    fromWallet.ID,
			ReferenceID: refID,
			Amount:      amount.Neg(),
			Currency:    fromWallet.Currency,
			Type:        entity.LedgerTypeDebit,
		}

		plusLedger := entity.Ledger{
			WalletID:    toWallet.ID,
			ReferenceID: refID,
			Amount:      amount,
			Currency:    toWallet.Currency,
			Type:        entity.LedgerTypeKredit,
		}

		if err := tx.Create(&minsLedger).Error; err != nil {
			w.Log.WithField("trace_id", traceID).WithError(err).Error("transfer: failed create ledger minus")
			return err
		}

		if err := tx.Create(&plusLedger).Error; err != nil {
			w.Log.WithField("trace_id", traceID).WithError(err).Error("transfer: failed create ledger plus")
			return err
		}

		return nil
	}); err != nil {
		return traceID, err
	}

	return traceID, nil
}
