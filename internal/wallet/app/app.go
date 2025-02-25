package app

import (
	logger "wallets/internal/logger/slog"
	"wallets/internal/wallet/config"
)

func Run() error {
	//TODO
	//init storage
	//init router
	//cleanenv

	config := config.MustLoad()

	log := logger.Init(config.Env)
	log.Info("Logger inited!")

	return nil
}
