package main

import "github.com/rodrigoTcarmo/inkion/pkg/mail"

func main() {
	emailClient := mail.NewClient()
	emailClient.GetEmails()
}
