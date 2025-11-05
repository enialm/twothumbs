// File: internal/integrations/email.go

// This file contains the welcome email template and the logic to send emails using SMTP.
// Both TLS (port 465) and STARTTLS (port 587) are supported.

package integrations

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/smtp"
	"strings"
	"time"

	"twothumbs/internal/models"
)

// *Email sending logic*

func NewEmailSender(host, port, username, password, from string) *models.EmailSender {
	return &models.EmailSender{
		SMTPHost: host,
		SMTPPort: port,
		Username: username,
		Password: password,
		From:     from,
	}
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func SendEmail(e *models.EmailSender, to, replyTo, subject, body string) error {
	headers := []string{
		"From: " + e.From,
		"To: " + to,
		"Reply-To: " + replyTo,
		"Subject: " + subject,
		"Date: " + time.Now().UTC().Format(time.RFC1123Z),
		"Message-ID: <" + randomString(12) + "@twothumbs.io>",
		"",
		body,
	}

	msg := []byte(strings.Join(headers, "\r\n"))
	addr := e.SMTPHost + ":" + e.SMTPPort

	switch e.SMTPPort {
	case "465":
		tlsconfig := &tls.Config{
			ServerName: e.SMTPHost,
		}
		conn, err := tls.Dial("tcp", addr, tlsconfig)
		if err != nil {
			return fmt.Errorf("tls dial error: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, e.SMTPHost)
		if err != nil {
			return fmt.Errorf("smtp new client error: %w", err)
		}
		defer client.Quit()

		auth := smtp.PlainAuth("", e.Username, e.Password, e.SMTPHost)

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth error: %w", err)
		}

		if err = client.Mail(e.From); err != nil {
			return fmt.Errorf("smtp mail error: %w", err)
		}

		if err = client.Rcpt(to); err != nil {
			return fmt.Errorf("smtp rcpt error: %w", err)
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("smtp data error: %w", err)
		}

		_, err = w.Write(msg)
		if err != nil {
			return fmt.Errorf("smtp write error: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("smtp close error: %w", err)
		}

		return nil

	default:
		// Default to 587 (STARTTLS)
		auth := smtp.PlainAuth("", e.Username, e.Password, e.SMTPHost)

		c, err := smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("smtp dial error: %w", err)
		}
		defer c.Quit()

		if err = c.StartTLS(&tls.Config{ServerName: e.SMTPHost}); err != nil {
			return fmt.Errorf("starttls error: %w", err)
		}

		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth error: %w", err)
		}

		if err = c.Mail(e.From); err != nil {
			return fmt.Errorf("smtp mail error: %w", err)
		}

		if err = c.Rcpt(to); err != nil {
			return fmt.Errorf("smtp rcpt error: %w", err)
		}

		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("smtp data error: %w", err)
		}

		_, err = w.Write(msg)
		if err != nil {
			return fmt.Errorf("smtp write error: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("smtp close error: %w", err)
		}

		return nil
	}
}
