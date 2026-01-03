package mail

import (
	"log/slog"
	"os"
)

func (m Mail) Auth() {
	email := os.Getenv("INKION_EMAIL_ADDRS")
	appPassword := os.Getenv("INKION_EMAIL_PWD")
	slog.Info("Authenticating to IMAP server")

	// Login with email and app password
	if err := m.client.Login(email, appPassword).Wait(); err != nil {
		slog.Error("Login failed", "error", err)
		return
	}

	if appPassword == "" || email == "" {
		slog.Error("INKION_EMAIL_PWD or INKION_EMAIL_ADDRS environment variable not set")
		return
	}

	slog.Info("Authentication successful!")
}
