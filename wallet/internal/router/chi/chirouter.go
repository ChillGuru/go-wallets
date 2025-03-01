package chirouter

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func InitWallet(r *chi.Mux) {
	r.Use(middleware.RequestID) //ручка для трейсинга запросов
	r.Use(middleware.Logger)    //ручка для логирования запросов
	r.Use(middleware.Recoverer) //ручка для отлова паник в других ручках

}
