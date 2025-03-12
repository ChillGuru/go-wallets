package chirouter

import (
	"wallet/internal/http/handlers"
	"wallet/internal/service"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func InitWallet(r *chi.Mux, s *service.WalletService) {

	r.Use(middleware.RequestID) //трейсинг запросов
	r.Use(middleware.Logger)    //логирование запросов
	r.Use(middleware.Recoverer) //отлов паник

	r.Post("/wallet", handlers.CreateWalletHandler(s))
	r.Get("/wallets", handlers.GetWalletsHandler(s))

	r.Route("/wallet", func(r chi.Router) {
		r.Delete("/{id}", handlers.RemoveWalletHandler(s))
	})

	r.Route("/wallets", func(r chi.Router) {
		r.Get("/{id}", handlers.GetWalletHandler(s))
		r.Put("/{id}", handlers.PutWalletsNameHandler(s))
		r.Post("/{id}/deposit", handlers.WalletDepositHandler(s))
		r.Post("/{id}/withdraw", handlers.WalletWithdrawHandler(s))
		r.Post("/{id}/transfer", handlers.WalletTransferHandler(s))
	})

}
