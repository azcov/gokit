package sendgrid

import (
	"context"
	"fmt"

	"github.com/azcov/gokit/email"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var _ email.Mailer = (*SendGrid)(nil)

type SendGrid struct {
	client *sendgrid.Client
}

func New(apiKey string) *SendGrid {
	return &SendGrid{client: sendgrid.NewSendClient(apiKey)}
}

func (s *SendGrid) Send(ctx context.Context, msg email.Message) error {
	m := mail.NewV3Mail()
	m.SetFrom(mail.NewEmail(msg.From.Name, msg.From.Email))
	m.Subject = msg.Subject

	p := mail.NewPersonalization()
	for _, a := range msg.To {
		p.AddTos(mail.NewEmail(a.Name, a.Email))
	}
	for _, a := range msg.CC {
		p.AddCCs(mail.NewEmail(a.Name, a.Email))
	}
	for _, a := range msg.BCC {
		p.AddBCCs(mail.NewEmail(a.Name, a.Email))
	}
	m.AddPersonalizations(p)

	if msg.HTML != "" {
		m.AddContent(mail.NewContent("text/html", msg.HTML))
	}
	if msg.Text != "" {
		m.AddContent(mail.NewContent("text/plain", msg.Text))
	}

	resp, err := s.client.SendWithContext(ctx, m)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("sendgrid: status %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

func (s *SendGrid) Close() error { return nil }
