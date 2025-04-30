package app

import (
	"log/slog"
	"net/http"
	"os"
	"stats/internal/config"
	"stats/internal/kafka"
	logger "stats/internal/logger/slog"
	chirouter "stats/internal/router/chi"
	"stats/internal/service"
	"stats/internal/storage/postgre"

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

	//Init kafka consumer
	kafkaConsumer, err := kafka.NewConsumer(config.Brokers, config.Topic, storage, log)
	if err != nil {
		log.Error("Can't init kafka consumer: ", logger.Err(err))
		os.Exit(1)
	}

	if kafkaConsumer != nil {
		log.Info("Kafka consumer inited")
	}

	// Init service
	statsService := service.New(storage)

	// Init router
	router := chi.NewRouter()
	chirouter.InitWallet(router, statsService)

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
