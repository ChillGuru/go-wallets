package chirouter

import (
	//"stats/internal/http/handlers"
	//"stats/internal/service"

	handler "stats/internal/http/handlers"
	"stats/internal/service"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func InitWallet(r *chi.Mux, s *service.StatsService) {

	r.Use(middleware.RequestID) //трейсинг запросов
	r.Use(middleware.Logger)    //логирование запросов
	r.Use(middleware.Recoverer) //отлов паник

	r.Get("/stats/wallets", handler.GetStatsHandler(s))
}
