package app

import (
	"log/slog"
	"net/http"
	"os"
	"wallet/internal/config"
	"wallet/internal/kafka"
	logger "wallet/internal/logger/slog"
	chirouter "wallet/internal/router/chi"
	"wallet/internal/service"
	"wallet/internal/storage/postgre"
	//"wallet/internal/storage/sqlite"
	"github.com/go-chi/chi"
)

func Run() error {
	//Load config
	config := config.MustLoad()

	//Init logger
	log := logger.Init(config.Env)
	log.Info("Logger inited!")

	//Init storage
	storage, err := postgre.New(config.DBServer.Host, config.DBServer.Port)
	if err != nil {
		log.Error("Can't init storage: ", logger.Err(err))
		os.Exit(1)
	}
	defer storage.Close()

	//Init kafka producer
	kafkaProducer, err := kafka.NewProducer(config.Brokers, config.Topic)
	if err != nil {
		log.Error("Can't init kafka producer: ", logger.Err(err))
		os.Exit(1)
	}

	//Init service
	walletService := service.New(storage, kafkaProducer)

	//Init router
	router := chi.NewRouter()
	chirouter.InitWallet(router, walletService)

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
