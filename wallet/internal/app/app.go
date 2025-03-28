package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"wallet/internal/config"
	logger "wallet/internal/logger/slog"
	chirouter "wallet/internal/router/chi"
	"wallet/internal/service"
	"wallet/internal/storage/sqlite"

	"github.com/go-chi/chi"
)

func Run() error {
	//TODO
	//init storage
	//cleanenv

	config := config.MustLoad()

	log := logger.Init(config.Env)
	log.Info("Logger inited!")

	storage, err := sqlite.New(config.StoragePath)
	if err != nil {
		log.Error("Can't init storage: ", logger.Err(err))
		os.Exit(1)
	}

	walletService := service.New(storage)

	router := chi.NewRouter()
	chirouter.InitWallet(router, walletService)

	wallets, err := storage.GetWallets(context.TODO())
	if err != nil {
		log.Error("Can't get wallets: ", logger.Err(err))
	}

	fmt.Printf("%+v\n", wallets)

	srv := &http.Server{
		Addr:         config.Address,
		Handler:      router,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
		IdleTimeout:  config.IdleTimeout,
	}

	log.Info("Staring server", slog.String("adress: ", config.Address))

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Can't run server: ", logger.Err(err))
	}

	log.Info("Wallet server stopped.")

	return nil
}
