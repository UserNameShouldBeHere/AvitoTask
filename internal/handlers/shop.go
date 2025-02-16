package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
	"go.uber.org/zap"
)

type CoinTransactionRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type ShopService interface {
	GetInfo(ctx context.Context, username string) (domain.InventoryInfo, error)
	SendCoin(ctx context.Context, transaction domain.Transaction) error
	BuyItem(ctx context.Context, username string, itemName string) error
}

type ShopHandler struct {
	authService AuthService
	shopService ShopService
	logger      *zap.SugaredLogger
}

func NewShopHandler(authService AuthService, shopService ShopService, logger *zap.SugaredLogger) (*ShopHandler, error) {
	return &ShopHandler{
		authService: authService,
		shopService: shopService,
		logger:      logger,
	}, nil
}

func (h *ShopHandler) Info(w http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: "",
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(customErrors.ErrUnauthenticated),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to unmarshall http request: %v", err)
		}
		return
	}

	ctx := context.WithValue(req.Context(), CtxSessionName, token.Value)

	name, ok := h.authService.GetNameAndCheck(ctx, token.Value)
	if !ok {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: token.Value,
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(customErrors.ErrUnauthenticated),
				Data:    ErrorResponse{Errors: customErrors.ErrUnauthenticated.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to unmarshall http request: %v", err)
		}
		return
	}

	ctx = context.WithValue(req.Context(), CtxSessionName, name)

	info, err := h.shopService.GetInfo(ctx, name)
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: name,
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to write http response: %v", err)
		}
		return
	}

	err = WriteResponse(
		w,
		h.logger,
		ResponseData{
			Session: name,
			Url:     req.Pattern,
			Status:  http.StatusOK,
			Data:    info,
		})
	if err != nil {
		h.logger.Errorf("unable to write http response: %v", err)
	}
}

func (h *ShopHandler) SendCoin(w http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: "",
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(customErrors.ErrUnauthenticated),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to unmarshall http request: %v", err)
		}
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: token.Value,
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to decode http request: %v", err)
		}
		return
	}

	var parsedReq CoinTransactionRequest
	err = json.Unmarshal(body, &parsedReq)
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: token.Value,
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to unmarshall http request: %v", err)
		}
		return
	}

	ctx := context.WithValue(req.Context(), CtxSessionName, token.Value)

	name, ok := h.authService.GetNameAndCheck(ctx, token.Value)
	if !ok {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: token.Value,
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(customErrors.ErrUnauthenticated),
				Data:    ErrorResponse{Errors: customErrors.ErrUnauthenticated.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to unmarshall http request: %v", err)
		}
		return
	}

	transaction := domain.Transaction{
		From:   name,
		To:     parsedReq.ToUser,
		Amount: parsedReq.Amount,
	}

	if err = transaction.Validate(); err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: token.Value,
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to unmarshall http request: %v", err)
		}
		return
	}

	ctx = context.WithValue(req.Context(), CtxSessionName, name)

	err = h.shopService.SendCoin(ctx, transaction)
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: token.Value,
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to write http response: %v", err)
		}
		return
	}

	err = WriteResponse(
		w,
		h.logger,
		ResponseData{
			Session: token.Value,
			Url:     req.Pattern,
			Status:  http.StatusOK,
			Data:    nil,
		})
	if err != nil {
		h.logger.Errorf("unable to write http response: %v", err)
	}
}

func (h *ShopHandler) BuyItem(w http.ResponseWriter, req *http.Request) {
	token, err := req.Cookie("token")
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: "",
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(customErrors.ErrUnauthenticated),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to unmarshall http request: %v", err)
		}
		return
	}

	ctx := context.WithValue(req.Context(), CtxSessionName, token.Value)

	name, ok := h.authService.GetNameAndCheck(ctx, token.Value)
	if !ok {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: token.Value,
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(customErrors.ErrUnauthenticated),
				Data:    ErrorResponse{Errors: customErrors.ErrUnauthenticated.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to unmarshall http request: %v", err)
		}
		return
	}

	itemName := req.PathValue("item")

	ctx = context.WithValue(req.Context(), CtxSessionName, name)

	err = h.shopService.BuyItem(ctx, name, itemName)
	if err != nil {
		err = WriteResponse(
			w,
			h.logger,
			ResponseData{
				Session: token.Value,
				Url:     req.Pattern,
				Status:  customErrors.ConvertToHttpErr(err),
				Data:    ErrorResponse{Errors: err.Error()},
			})
		if err != nil {
			h.logger.Errorf("unable to write http response: %v", err)
		}
		return
	}

	err = WriteResponse(
		w,
		h.logger,
		ResponseData{
			Session: name,
			Url:     req.Pattern,
			Status:  http.StatusOK,
			Data:    nil,
		})
	if err != nil {
		h.logger.Errorf("unable to write http response: %v", err)
	}
}
