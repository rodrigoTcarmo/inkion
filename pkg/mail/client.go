package mail

import (
	"log/slog"
	"os"

	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/rodrigoTcarmo/inkion/pkg/apis/config"
	v1 "k8s.io/api/core/v1"
)

type Client interface {
	Auth()
	FetchEmails()
}

type Mail struct {
	client *imapclient.Client
	config *config.InkionConfig
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
	return newConfig(client)
}

func newConfig(client *imapclient.Client) *Mail {
	inkionConfig, err := config.NewInkionFromConfigMap(&v1.ConfigMap{
		Data: map[string]string{
			"expected-sender": os.Getenv("INKION_EXPECTED_SENDER"), //todo: convert this to read actual configMaps from the cluster
		},
	})
	if err != nil {
		slog.Error("failed to load configs", "error", err)
		return nil
	}
	return &Mail{client: client, config: inkionConfig}
}

func Close(client *imapclient.Client) {
	client.Close()
}
