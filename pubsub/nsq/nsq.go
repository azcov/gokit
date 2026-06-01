package nsq

import (
	"context"

	"github.com/azcov/gokit/pubsub"
	gonsq "github.com/nsqio/go-nsq"
)

var _ pubsub.PubSub = (*NSQ)(nil)

type NSQ struct {
	producer    *gonsq.Producer
	consumers   []*gonsq.Consumer
	cfg         *gonsq.Config
	nsqdAddr    string
	lookupdAddr string
}

type Config struct {
	NSQDAddr    string       `env:"NSQD_ADDRESS" json:"nsqd_address" yaml:"nsqd_address"`
	LookupdAddr string       `env:"LOOKUPD_ADDRESS" json:"lookupd_address" yaml:"lookupd_address"`
	Config      *gonsq.Config `json:"-" yaml:"-"`
}

func New(cfg Config) (*NSQ, error) {
	c := cfg.Config
	if c == nil {
		c = gonsq.NewConfig()
	}
	p, err := gonsq.NewProducer(cfg.NSQDAddr, c)
	if err != nil {
		return nil, err
	}
	return &NSQ{
		producer:    p,
		cfg:         c,
		nsqdAddr:    cfg.NSQDAddr,
		lookupdAddr: cfg.LookupdAddr,
	}, nil
}

func (n *NSQ) Publish(_ context.Context, topic string, msg pubsub.Message) error {
	return n.producer.Publish(topic, msg.Payload)
}

func (n *NSQ) Subscribe(ctx context.Context, topic string, handler pubsub.Handler) error {
	c, err := gonsq.NewConsumer(topic, "default", n.cfg)
	if err != nil {
		return err
	}
	c.AddHandler(gonsq.HandlerFunc(func(msg *gonsq.Message) error {
		return handler(ctx, pubsub.Message{Topic: topic, Payload: msg.Body})
	}))
	if n.lookupdAddr != "" {
		err = c.ConnectToNSQLookupd(n.lookupdAddr)
	} else {
		err = c.ConnectToNSQD(n.nsqdAddr)
	}
	if err != nil {
		return err
	}
	n.consumers = append(n.consumers, c)
	return nil
}

func (n *NSQ) Unsubscribe(_ string) error { return nil }

func (n *NSQ) Close() error {
	n.producer.Stop()
	for _, c := range n.consumers {
		c.Stop()
	}
	return nil
}
