package main

import (
	"encoding/json"
	"log/slog"

	"github.com/rodrigoTcarmo/inkion/pkg/mail"
	"github.com/rodrigoTcarmo/inkion/pkg/transaction"
)

func main() {
	emailClient := mail.NewClient()

	tr, err := transaction.BuildTransaction(emailClient)
	if err != nil {
		slog.Error("error building the transaction", "error", err)
	}

	pretty, err := json.MarshalIndent(tr, "", "  ")
	if err != nil {
		slog.Error("error marshaling transaction to json", "error", err)
	} else {
		println(string(pretty))
	}
}
