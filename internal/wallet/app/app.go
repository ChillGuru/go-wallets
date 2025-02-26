package app

import (
	"net/http"
	logger "wallets/internal/logger/slog"
	"wallets/internal/wallet/config"

	"github.com/go-chi/chi"
)

func Run() error {
	//TODO
	//init storage
	//cleanenv

	config := config.MustLoad()

	log := logger.Init(config.Env)
	log.Info("Logger inited!")

	router := chi.NewRouter()

	srv := &http.Server{
		Addr:         config.Address,
		Handler:      router,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
		IdleTimeout:  config.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Can't run server: ", logger.Err(err))
	}

	log.Info("Wallet server stopped.")

	return nil
}
