package email

import "context"

type Address struct {
	Name  string
	Email string
}

type Message struct {
	From    Address
	To      []Address
	CC      []Address
	BCC     []Address
	ReplyTo *Address
	Subject string
	HTML    string
	Text    string
}

type Mailer interface {
	Send(ctx context.Context, msg Message) error
	Close() error
}
