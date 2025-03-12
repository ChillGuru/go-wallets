package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type WalletReplenisher interface {
	Deposit(ctx context.Context, walletID string, amount float64) (int64, error)
}

func WalletDepositHandler(replenisher WalletReplenisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		walletID := chi.URLParam(r, "id")
		if walletID == "" {
			render.JSON(w, r, Error("Invalid request"))
			return
		}

		var req Request

		//разбираем запрос
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			render.JSON(w, r, Error("Can't decode request"))
			return
		}
		if req.Amount <= 0 {
			render.JSON(w, r, Error("Deposit amount must be more than 0"))
			return
		}

		//валидируем запрос
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			render.JSON(w, r, ValidationError(validateErr))
			return
		}

		_, err := replenisher.Deposit(r.Context(), walletID, req.Amount)
		if err != nil {
			render.JSON(w, r, Error(err.Error()))
			return
		}

		render.JSON(w, r, Response{
			Success: true,
			Amount:  req.Amount,
		})
	}
}

type WalletWithdrawer interface {
	Withdraw(ctx context.Context, walletID string, amount float64) (int64, error)
}

func WalletWithdrawHandler(withdrawer WalletWithdrawer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		walletID := chi.URLParam(r, "id")
		if walletID == "" {
			render.JSON(w, r, Error("Invalid request"))
			return
		}

		var req Request

		//разбираем запрос
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			render.JSON(w, r, Error("Can't decode request"))
			return
		}
		if req.Amount <= 0 {
			render.JSON(w, r, Error("Deposit amount must be more than 0"))
			return
		}

		//валидируем запрос
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			render.JSON(w, r, ValidationError(validateErr))
			return
		}

		_, err := withdrawer.Withdraw(r.Context(), walletID, req.Amount)
		if err != nil {
			render.JSON(w, r, Error(err.Error()))
			return
		}

		render.JSON(w, r, Response{
			Success: true,
			Amount:  req.Amount,
		})
	}
}

type WalletTransferer interface {
	Transfer(ctx context.Context, walletID string, amount float64, transferTo string) (int64, int64, error)
}

func WalletTransferHandler(transferer WalletTransferer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		walletID := chi.URLParam(r, "id")
		if walletID == "" {
			render.JSON(w, r, Error("Invalid request"))
			return
		}

		var req Request

		//разбираем запрос
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			render.JSON(w, r, Error("Can't decode request"))
			return
		}
		if req.Amount <= 0 {
			render.JSON(w, r, Error("Deposit amount must be more than 0"))
			return
		}
		if req.TransferTo == "" {
			render.JSON(w, r, Error("Empty reciever wallet ID"))
			return
		}

		//валидируем запрос
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)
			render.JSON(w, r, ValidationError(validateErr))
			return
		}

		_, _, err := transferer.Transfer(r.Context(), walletID, req.Amount, req.TransferTo)
		if err != nil {
			render.JSON(w, r, Error(err.Error()))
			return
		}

		render.JSON(w, r, Response{
			Success: true,
			Amount:  req.Amount,
		})
	}
}
