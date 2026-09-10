package routes

import (
	"transfer-exam/internal/controller"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(wallet controller.WalletController) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
	}))

	v1 := r.Group("/api/v1")
	{

		wal := v1.Group("/wallets")
		{
			wal.POST("/transfer", wallet.TransferWallet)
		}

	}

	return r
}
