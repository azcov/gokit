package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/azcov/gokit/pubsub"
)

var _ pubsub.PubSub = (*Kafka)(nil)

type Kafka struct {
	producer sarama.SyncProducer
	consumer sarama.ConsumerGroup
}

type Config struct {
	Brokers []string       `env:"BROKERS" json:"brokers" yaml:"brokers"`
	GroupID string         `env:"GROUP_ID" json:"group_id" yaml:"group_id"`
	Config  *sarama.Config `json:"-" yaml:"-"`
}

func New(cfg Config) (*Kafka, error) {
	c := cfg.Config
	if c == nil {
		c = sarama.NewConfig()
		c.Producer.Return.Successes = true
	}
	producer, err := sarama.NewSyncProducer(cfg.Brokers, c)
	if err != nil {
		return nil, err
	}
	consumer, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, c)
	if err != nil {
		producer.Close()
		return nil, err
	}
	return &Kafka{producer: producer, consumer: consumer}, nil
}

func (k *Kafka) Publish(_ context.Context, topic string, msg pubsub.Message) error {
	_, _, err := k.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(msg.Payload),
	})
	return err
}

func (k *Kafka) Subscribe(ctx context.Context, topic string, handler pubsub.Handler) error {
	h := &consumerHandler{handler: handler}
	go func() {
		for {
			if ctx.Err() != nil {
				return
			}
			_ = k.consumer.Consume(ctx, []string{topic}, h)
		}
	}()
	return nil
}

func (k *Kafka) Unsubscribe(_ string) error { return nil }

func (k *Kafka) Close() error {
	if err := k.producer.Close(); err != nil {
		return err
	}
	return k.consumer.Close()
}

type consumerHandler struct {
	handler pubsub.Handler
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if err := h.handler(sess.Context(), pubsub.Message{
			Topic:   msg.Topic,
			Payload: msg.Value,
		}); err != nil {
			return err
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
