package handlers

import (
	"context"
	"errors"
	"net/http"
	"wallet/internal/storage"

	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type WalletCreator interface {
	CreateWallet(ctx context.Context, name string) (*storage.Wallet, error)
}

func CreateWalletHandler(creator WalletCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.CreateWallet"

		var req Request

		//разбираем запрос
		if err := render.DecodeJSON(r.Body, &req); err != nil {
			render.JSON(w, r, Error("Can't decode request"))
			return
		}

		//валидируем запрос
		if err := validator.New().Struct(req); err != nil {
			validateErr := err.(validator.ValidationErrors)

			render.JSON(w, r, ValidationError(validateErr))

			return
		}

		//создаем кашелек
		createdWallet, err := creator.CreateWallet(r.Context(), req.Name)
		if errors.Is(err, storage.ErrWalletExists) {
			render.JSON(w, r, Error("Can't create wallet. Wallet already exists"))
			return
		}
		if err != nil {
			render.JSON(w, r, Error("Can't create wallet"))
			return
		}

		render.JSON(w, r, Response{
			ID:     createdWallet.ID,
			Name:   createdWallet.Name,
			Status: createdWallet.Status,
		})
	}
}
