// Package pubsub defines publish/subscribe interfaces (Publisher, Subscriber).
// Messages fan out to all subscribers of a topic.
//
// Implementations: pubsub/memory, pubsub/kafka, pubsub/nsq.
package pubsub
