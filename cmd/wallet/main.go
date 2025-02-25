package main

import (
	"log"
	"wallets/internal/wallet/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
