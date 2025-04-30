package handler

import (
	"context"
	"log"
	"net/http"
	"stats/internal/storage"

	"github.com/go-chi/render"
)

type Response struct {
	Total      int     `json:"total"`
	Active     int     `json:"active"`
	Inactive   int     `json:"inactive"`
	Deposited  float64 `json:"deposited"`
	Withdrawn  float64 `json:"withdrawn"`
	Transfered float64 `json:"transfered"`
	ErrCode    string  `json:"err_code,omitempty"`
}

type StatsRecipient interface {
	GetStats(ctx context.Context) (*storage.Stats, error)
}

func GetStatsHandler(recipient StatsRecipient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		stats, err := recipient.GetStats(r.Context())
		if err != nil {
			render.JSON(w, r, Response{ErrCode: err.Error()})
			return
		}

		log.Print(stats.Total)

		render.JSON(w, r, Response{
			Total:      stats.Total,
			Active:     stats.Active,
			Inactive:   stats.Inactive,
			Deposited:  stats.Deposited,
			Withdrawn:  stats.Withdrawn,
			Transfered: stats.Transfered,
		})
	}
}
