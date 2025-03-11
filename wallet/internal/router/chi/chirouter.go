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
	r.Get("/wallets/{id}", handlers.GetWalletHandler(s))
	r.Get("/wallets", handlers.GetWalletsHandler(s))
	r.Put("/wallets/{id}", handlers.PutWalletsNameHandler(s))
	r.Delete("/wallet/{id}", handlers.RemoveWalletHandler(s))
	//r.Post("/wallets/{id}/deposit", )
	//r.Post("/wallets/{id}/withdraw", )
	//r.Post("/wallets/{id}/transfer", )
}
