package smtp

import (
	"bytes"
	"context"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/azcov/gokit/email"
)

var _ email.Mailer = (*SMTP)(nil)

type SMTP struct {
	host     string
	port     int
	username string
	password string
}

type Config struct {
	Host     string `config:"host"`
	Port     int    `config:"port"`
	Username string `config:"username"`
	Password string `config:"password"`
}

func New(cfg Config) *SMTP {
	return &SMTP{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
	}
}

func (s *SMTP) Send(_ context.Context, msg email.Message) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	to := make([]string, len(msg.To))
	for i, a := range msg.To {
		to[i] = a.Email
	}

	body := buildMIME(msg, to)
	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	return smtp.SendMail(addr, auth, msg.From.Email, to, body)
}

func (s *SMTP) Close() error { return nil }

func buildMIME(msg email.Message, to []string) []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s <%s>\r\n", msg.From.Name, msg.From.Email))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	buf.WriteString("MIME-version: 1.0\r\n")
	if msg.HTML != "" {
		buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
		buf.WriteString(msg.HTML)
	} else {
		buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
		buf.WriteString(msg.Text)
	}
	return buf.Bytes()
}
