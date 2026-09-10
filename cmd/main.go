package main

import (
	"transfer-exam/infrastructure/config"
	"transfer-exam/infrastructure/datastore"
	"transfer-exam/infrastructure/logger"
	"transfer-exam/internal/controller"
	"transfer-exam/internal/repository"
	"transfer-exam/internal/routes"
	"transfer-exam/internal/service"

	"github.com/go-playground/validator/v10"
)

func main() {
	config, err := config.NewViper()
	if err != nil {
		panic(err)
	}

	log := logger.NewLogrus(&config.Logger)
	db := datastore.NewDatabase(&config.Postgres)
	datastore.NewRedis(&config.Redis)
	validate := validator.New()
	
	//wallet
	walletRepo := repository.NewWalletRepositoryImpl()
	walletService := service.NewWalletServiceImpl(walletRepo, db, log, validate)
	walletController := controller.NewWalletControllerImpl(walletService)


	router := routes.NewRouter(walletController)

	router.Run(config.App.Port)
}
