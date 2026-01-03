package mail

import (
	"log/slog"
	"os"

	"github.com/emersion/go-imap/v2/imapclient"
)

type Client interface {
	Auth()
	FetchEmails()
}

func NewClient() *Mail {
	imapServer := os.Getenv("INKION_IMAP_SERVER")

	slog.Info("Connecting to imap server", "IMAP Server", imapServer)

	// Connect to Gmail IMAP server
	client, err := imapclient.DialTLS(imapServer, nil)
	if err != nil {
		slog.Error("failed to connect", "error", err)
		return nil
	}

	slog.Info("Connected!")
	return &Mail{
		client: client,
	}
}

func Close(client *imapclient.Client) {
	client.Close()
}
