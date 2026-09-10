package response

import (
	"errors"
	"net/http"
	"transfer-exam/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type WebResponse struct {
	Success     bool       `json:"success"`
	TraceID     *string     `json:"trace_id,omitempty"`
	ErrorCode   *errorCode `json:"error_code,omitempty"`
	Message     string     `json:"message"`
	Severity    *string    `json:"severity,omitempty"`
	IsRetryable bool       `json:"is_retryable"`
	Data        *any       `json:"data,omitempty"`
}

type ErrorHttpResp struct {
	Code int
	Resp *WebResponse
}

type errorCode string

var (
	ErrCodeAmountInvalid       errorCode = "AMOUNT_MUST_BE_GREATHER_THAN_ZERO"
	ErrCodeAccountInvalid      errorCode = "CANNOT_TRANSFER_SAME_ACCOUNT"
	ErrCodeCurrencyInvalid     errorCode = "CURRENCY_INVALID"
	ErrCodeInsufficientBalance errorCode = "INSUFFICIENT_BALANCE"
	ErrCodeInternalServer      errorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeRecordNotFound      errorCode = "RECORD_NOT_FOUND"

	SeverityError string = "ERROR"
	SeverityWarning string = "WARNING"
	SeverityInfo    string = "INFO"
)

func SuccessResponse(data any,message ...string) *WebResponse {
	resMessage := "Success"
	if len(message) > 0 {
		resMessage = message[0]
	}
	return &WebResponse{
		Success: true,
		Message: resMessage,
		Data:    &data,
	}
}

func ErrorValidate(message ...string) *WebResponse {
	resMessage := "Error"
	if len(message) > 0 {
		resMessage = message[0]
	}
	return &WebResponse{
		Success: false,
		Message: resMessage,
		Data:    nil,
	}
}

func ErrorResponse(err error,traceID string) *ErrorHttpResp {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return &ErrorHttpResp{
			Code: http.StatusNotFound,
			Resp: &WebResponse{
				Success:     false,
				TraceID:     &traceID,
				ErrorCode:   &ErrCodeRecordNotFound,
				Message:     err.Error(),
				Severity:    &SeverityWarning,
				IsRetryable: false,
			},
		}
	case errors.Is(err, service.ErrAmountMustBeGreater):
		return &ErrorHttpResp{
			Code: http.StatusBadRequest,
			Resp: &WebResponse{
				Success:     false,
				TraceID:     &traceID,
				ErrorCode:   &ErrCodeAmountInvalid,
				Message:     err.Error(),
				Severity:    &SeverityWarning,
				IsRetryable: false,
			},
		}
	case errors.Is(err, service.ErrCannotSameAccount):
		return &ErrorHttpResp{
			Code: http.StatusBadRequest,
			Resp: &WebResponse{
				Success:     false,
				TraceID:     &traceID,
				ErrorCode:   &ErrCodeAccountInvalid,
				Message:     err.Error(),
				Severity:    &SeverityWarning,
				IsRetryable: false,
			},
		}

	case errors.Is(err, service.ErrInvalidCurrency):
		return &ErrorHttpResp{
			Code: http.StatusBadRequest,
			Resp: &WebResponse{
				Success:     false,
				TraceID:     &traceID,
				ErrorCode:   &ErrCodeCurrencyInvalid,
				Message:     err.Error(),
				Severity:    &SeverityWarning,
				IsRetryable: false,
			},
		}

	case errors.Is(err, service.ErrInsufficientBalance):
		return &ErrorHttpResp{
			Code: http.StatusBadRequest,
			Resp: &WebResponse{
				Success:     false,
				TraceID:     &traceID,
				ErrorCode:   &ErrCodeInsufficientBalance,
				Message:     err.Error(),
				Severity:    &SeverityWarning,
				IsRetryable: false,
			},
		}

	default:
		return &ErrorHttpResp{
			Code: http.StatusInternalServerError,
			Resp: &WebResponse{
				Success:     false,
				TraceID:     &traceID,
				ErrorCode:   &ErrCodeInternalServer,
				Message:     "Internal Server Error",
				Severity:    &SeverityError,
				IsRetryable: false,
			},
		}
	}

}

func ErrorResponseJSON(c *gin.Context, traceID string, err error) *ErrorHttpResp {
	result := ErrorResponse(err, traceID)
	c.JSON(result.Code, result.Resp)
	return result
}
