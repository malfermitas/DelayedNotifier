package sender

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"
)

type EmailSenderConfig struct {
	SMTPHost string
	SMTPPort int
	SMTPUser string
	SMTPPass string
	From     string
}

type EmailSender struct {
	cfg EmailSenderConfig
}

type EmailMessage struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
	From    string `json:"from,omitempty"`
}

func NewEmailSender(cfg EmailSenderConfig) (*EmailSender, error) {
	if cfg.SMTPHost == "" {
		return nil, errors.New("smtp host is required")
	}
	if cfg.SMTPPort == 0 {
		return nil, errors.New("smtp port is required")
	}
	if cfg.From == "" {
		return nil, errors.New("from address is required")
	}

	return &EmailSender{cfg: cfg}, nil
}

func (s *EmailSender) Send(ctx context.Context, toEmail string, text string) error {
	address := net.JoinHostPort(s.cfg.SMTPHost, strconv.Itoa(s.cfg.SMTPPort))

	headers := make([]string, 0, 5)
	headers = append(headers, "From: "+s.cfg.From)
	headers = append(headers, "To: "+toEmail)
	headers = append(headers, "Subject: Notification")
	headers = append(headers, "MIME-Version: 1.0")
	headers = append(headers, "Content-Type: text/plain; charset=UTF-8")

	payload := strings.Join(headers, "\r\n") + "\r\n\r\n" + text

	var auth smtp.Auth
	if s.cfg.SMTPUser != "" || s.cfg.SMTPPass != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUser, s.cfg.SMTPPass, s.cfg.SMTPHost)
	}

	if err := smtp.SendMail(address, auth, s.cfg.From, []string{toEmail}, []byte(payload)); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	return nil
}
