package repository

import (
	"transfer-exam/internal/entity"
)

type WalletRepository interface {
	Repository[entity.Wallet]
}

type walletRepositoryImpl struct {
	repositoryImpl[entity.Wallet]
}

func NewWalletRepositoryImpl() WalletRepository {
	return &walletRepositoryImpl{}
}
