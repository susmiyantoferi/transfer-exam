package controller

import (
	"net/http"
	"transfer-exam/internal/dto"
	"transfer-exam/internal/response"
	"transfer-exam/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletController interface {
	TransferWallet(c *gin.Context)
}

type walletControllerImpl struct {
	WalletService service.WalletService
}

func NewWalletControllerImpl(walletService service.WalletService) WalletController {
	return &walletControllerImpl{
		WalletService: walletService,
	}
}

func (w *walletControllerImpl) TransferWallet(c *gin.Context) {
	var req dto.TransferWalletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorValidate("invalid request body"))
		return
	}

	traceID, err := w.WalletService.TransferWallet(c.Request.Context(), &req)
	if err != nil {
		response.ErrorResponseJSON(c, traceID.String(), err)
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse(nil, "ok"))
}
