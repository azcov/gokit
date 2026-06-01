package resend

import (
	"context"
	"fmt"

	"github.com/azcov/gokit/email"
	"github.com/resend/resend-go/v2"
)

var _ email.Mailer = (*Resend)(nil)

type Resend struct {
	client *resend.Client
}

func New(apiKey string) *Resend {
	return &Resend{client: resend.NewClient(apiKey)}
}

func (r *Resend) Send(_ context.Context, msg email.Message) error {
	to := make([]string, len(msg.To))
	for i, a := range msg.To {
		to[i] = formatAddr(a)
	}
	_, err := r.client.Emails.Send(&resend.SendEmailRequest{
		From:    formatAddr(msg.From),
		To:      to,
		Subject: msg.Subject,
		Html:    msg.HTML,
		Text:    msg.Text,
	})
	return err
}

func (r *Resend) Close() error { return nil }

func formatAddr(a email.Address) string {
	if a.Name != "" {
		return fmt.Sprintf("%s <%s>", a.Name, a.Email)
	}
	return a.Email
}
