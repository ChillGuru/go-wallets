package handlers

import (
	"context"
	"errors"
	"net/http"
	"wallet/internal/storage"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
)

type WalletCreator interface {
	CreateWallet(ctx context.Context, name string) (*storage.Wallet, error)
}

func CreateWalletHandler(creator WalletCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.CreateWalletHandler"

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

type WalletRecipient interface {
	GetWallet(ctx context.Context, walletID string) (*storage.Wallet, error)
}

func GetWalletHandler(recipient WalletRecipient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		walletID := chi.URLParam(r, "id")
		if walletID == "" {
			render.JSON(w, r, Error("Invalid request"))
			return
		}

		wallet, err := recipient.GetWallet(r.Context(), walletID)
		if errors.Is(err, storage.ErrWalletNotExist) {
			render.JSON(w, r, Error("Wallet not exists"))
			return
		}
		if err != nil {
			render.JSON(w, r, Error("Can't get wallet"))
			return
		}

		render.JSON(w, r, wallet)
	}
}

type WalletsRecipient interface {
	GetWallets(ctx context.Context) ([]storage.Wallet, error)
}

func GetWalletsHandler(recipient WalletsRecipient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		wallets, err := recipient.GetWallets(r.Context())
		if err != nil {
			render.JSON(w, r, Error("Can't get list of wallets"))
			return
		}

		render.JSON(w, r, wallets)
	}
}

type WalletRenamer interface {
	UpdateName(ctx context.Context, walletID, name string) (int64, error)
}

func PutWalletsNameHandler(renamer WalletRenamer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		walletID := chi.URLParam(r, "id")
		if walletID == "" {
			render.JSON(w, r, Response{ID: walletID, Success: false, ErrCode: "Invalid request"})
			return
		}

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			render.JSON(w, r, Response{ID: walletID, Success: false, ErrCode: err.Error()})
			return
		}

		if req.Name == "" || len(req.Name) <= 1 {
			render.JSON(w, r, Response{ID: walletID, Success: false, ErrCode: "The name length must be more than 1 character"})
			return
		}

		_, err := renamer.UpdateName(r.Context(), walletID, req.Name)
		if err != nil {
			render.JSON(w, r, Response{ID: walletID, Success: false, ErrCode: err.Error()})
			return
		}

		render.JSON(w, r, Response{ID: walletID, Success: true})
	}
}
