package queue

import (
	"context"
)

type QueueType int

func (q QueueType) String() string {
	if q == QueueTypeAtLeastOnce {
		return "at-least-once"
	} else if q == QueueTypeAtMostOnce {
		return "at-most-once"
	}
	return ""
}

// delivery guaranties
const (
	QueueTypeAtLeastOnce = iota
	QueueTypeAtMostOnce
)

// Config queue configuration
type Config struct {
	Host      string
	Port      string
	ClusterId string
}

// Queue allows async communication with a message queue
type Queue interface {
	// Open opens connection
	// clientId must be unique
	Open(ctx context.Context, clientId string, options *Config) error
	// Close closes connection
	Close() error
	// Publish publishes a message to topic
	Publish(ctx context.Context, qt QueueType, topic string, msg *Message) error
	// Subscribe subscribes on topic
	Subscribe(qt QueueType, topic string, receiverChan chan<- []byte) error
	// SubscribeLB subscribes on topic with load balancing
	// if more than one subscribers specify the same loadBalancingGroup, messages are balanced among all subscribers within the group
	// so that the only one subscriber gets the message
	SubscribeLB(qt QueueType, topic, loadBalancingGroup string, receiverChan chan<- []byte) error
}
