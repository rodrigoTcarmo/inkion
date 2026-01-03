package mail

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/emersion/go-message/mail"
)

// Email represents a parsed email message
type Email struct {
	SeqNum      uint32
	Subject     string
	From        string
	To          string
	Date        string
	Body        string
	HTMLBody    string
	Attachments []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Size        int
}

func GetEmails() {
	// Gmail IMAP settings
	imapServer := "imap.gmail.com:993"
	email := os.Getenv("INKION_EMAIL_ADDRS")
	appPassword := os.Getenv("INKION_EMAIL_PWD")

	if appPassword == "" || email == "" {
		slog.Error("INKION_EMAIL_PWD or INKION_EMAIL_ADDRS environment variable not set")
		return
	}

	slog.Info("Connecting to imap server", "IMAP Server", imapServer)

	// Connect to Gmail IMAP server
	client, err := imapclient.DialTLS(imapServer, nil)
	if err != nil {
		slog.Error("failed to connect", "error", err)
		return
	}
	defer client.Close()

	fmt.Println("Connected! Logging in...")

	// Login with email and app password
	if err := client.Login(email, appPassword).Wait(); err != nil {
		slog.Error("Login failed", "error", err)
		return
	}

	fmt.Println("Login successful!")

	// Select INBOX
	mbox, err := client.Select("INBOX", nil).Wait()
	if err != nil {
		slog.Error("Failed to select INBOX", "error", err)
		return
	}

	fmt.Printf("\nINBOX has %d messages\n", mbox.NumMessages)

	// Fetch the last 5 emails (or fewer if inbox has less)
	numToFetch := uint32(5)
	if mbox.NumMessages < numToFetch {
		numToFetch = mbox.NumMessages
	}

	if numToFetch == 0 {
		slog.Info("No emails to fetch")
		return
	}

	// Create sequence set for the last N messages
	startSeq := mbox.NumMessages - numToFetch + 1
	var seqSet imap.SeqSet
	seqSet.AddRange(startSeq, mbox.NumMessages)

	// Fetch options - get envelope (headers) and body
	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		BodySection: []*imap.FetchItemBodySection{{}}, // Fetch full body
	}

	slog.Info("Fetching emails...", "emails quantity", numToFetch)

	fetchCmd := client.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	emails := []Email{}

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		parsedEmail := parseEmail(msg)
		emails = append(emails, parsedEmail)
	}

	if err := fetchCmd.Close(); err != nil {
		slog.Error("Fetch failed", "error", err)
	}

	// Display the emails
	fmt.Printf("\n========================================\n")
	fmt.Printf("       FETCHED %d EMAILS\n", len(emails))
	fmt.Printf("========================================\n")

	for i, e := range emails {
		fmt.Printf("\n--- Email %d ---\n", i+1)
		fmt.Printf("Subject: %s\n", e.Subject)
		fmt.Printf("From:    %s\n", e.From)
		fmt.Printf("To:      %s\n", e.To)
		fmt.Printf("Date:    %s\n", e.Date)

		if len(e.Attachments) > 0 {
			fmt.Printf("Attachments (%d):\n", len(e.Attachments))
			for _, att := range e.Attachments {
				fmt.Printf("  - %s (%s)\n", att.Filename, att.ContentType)
			}
		}

		// Show body preview (first 200 chars)
		fmt.Printf("Body preview:\n%s\n", e.Body)
	}
}

func parseEmail(msg *imapclient.FetchMessageData) Email {
	email := Email{
		SeqNum: msg.SeqNum,
	}

	// Iterate through all the data items
	for {
		item := msg.Next()
		if item == nil {
			break
		}

		switch data := item.(type) {
		case imapclient.FetchItemDataEnvelope:
			// Parse envelope (headers)
			env := data.Envelope
			if env != nil {
				email.Subject = env.Subject
				email.Date = env.Date.Format("2006-01-02 15:04:05")

				if len(env.From) > 0 {
					from := env.From[0]
					if from.Name != "" {
						email.From = fmt.Sprintf("%s <%s@%s>", from.Name, from.Mailbox, from.Host)
					} else {
						email.From = fmt.Sprintf("%s@%s", from.Mailbox, from.Host)
					}
				}

				if len(env.To) > 0 {
					to := env.To[0]
					if to.Name != "" {
						email.To = fmt.Sprintf("%s <%s@%s>", to.Name, to.Mailbox, to.Host)
					} else {
						email.To = fmt.Sprintf("%s@%s", to.Mailbox, to.Host)
					}
				}
			}

		case imapclient.FetchItemDataBodySection:
			// Parse the email body using go-message
			if data.Literal == nil {
				continue
			}

			mr, err := mail.CreateReader(data.Literal)
			if err != nil {
				continue
			}

			// Read each part
			for {
				part, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue
				}

				switch h := part.Header.(type) {
				case *mail.InlineHeader:
					// This is the email body (text or HTML)
					contentType, _, _ := h.ContentType()
					body, _ := io.ReadAll(part.Body)

					if strings.HasPrefix(contentType, "text/plain") {
						email.Body = string(body)
					} else if strings.HasPrefix(contentType, "text/html") {
						email.HTMLBody = string(body)
					}

				case *mail.AttachmentHeader:
					// This is an attachment
					filename, _ := h.Filename()
					contentType, _, _ := h.ContentType()
					body, _ := io.ReadAll(part.Body)

					email.Attachments = append(email.Attachments, Attachment{
						Filename:    filename,
						ContentType: contentType,
						Size:        len(body),
					})
				}
			}
		}
	}

	return email
}
