package messaging

import (
	"context"
	"time"
)

type HostConfig struct{
	Address string
	Qos int
}

func (h *HostConfig) GetQos() int{
	if h.Qos == 0{
		return 10
	}

	return h.Qos
}

// Host is the container which is used
// to host all consumers that are registered.
// It is responsible for the amqp connection
// starting & gracefully stopping all running consumers
// h := NewRabbitHost().Init(cfg.Host)
// h.AddBroker(NewBroker(cfg.Exchange, [])
type Host interface{
	// Init sets up the initial connection & quality of service
	// to be used by all registered consumers
	Init(context.Context, *HostConfig) (err error)
	// AddBroker will register an exchange and n consumers
	// which will consume from that exchange
	AddBroker(context.Context, *ExchangeConfig, []Consumer) error
	// Start will setup all queues and routing keys
	// assigned to each consumer and then in turn start them
	Run(context.Context) (err error)

	// Stop can be called when you wish to shut down the host
	Stop(context.Context) error
}

// AMQPMessage is a generic AMQPMessage message that
// can get transformed via middleware
// to something more useful
type AMQPMessage struct{
	ID string
	UserID string
	Subject string
	ReplyTo string
	CorrelationID string
	ContentType string
	ContentEncoding string
	Expiry string
	Created time.Time
	Headers map[string]interface{}
	Body interface{}
	Args map[string]interface{}
}

type MessageHandler interface {
	HandleMessage(ctx context.Context, m AMQPMessage) error
}

type HandlerFunc func(context.Context, AMQPMessage) error

func (f HandlerFunc) HandleMessage(ctx context.Context, m AMQPMessage) error {
	return f(ctx, m)
}

