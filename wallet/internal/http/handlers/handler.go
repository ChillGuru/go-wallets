package handlers

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator"
)

type Response struct {
	ID      string  `json:"id"`
	Name    string  `json:"name,omitempty"`
	Balance float64 `json:"balance,omitempty"`
	Status  string  `json:"status,omitempty"`
	Amount  float64 `json:"amount,omitempty"`
	Success bool    `json:"success,omitempty"`
	ErrCode string  `json:"err_code,omitempty"`
}

type Request struct {
	Amount     float64 `json:"amount,omitempty"`
	Name       string  `json:"name,omitempty"`
	TransferTo float64 `json:"transfer_to,omitempty"`
}

func Error(msg string) Response {
	return Response{
		ErrCode: msg,
	}
}

// Обработчик ошибок валидатора. Опционально, реализовано для повышения читаемости ошибок
// TODO: Добавить больше кейсов
func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		ErrCode: strings.Join(errMsgs, ", "),
	}
}
