package consumer

import (
	"context"
	"github.com/streadway/amqp"
)

// MessageHandler works in the same way httpHandlers do
// allowing middleware etc to be used on consumers
type MessageHandler interface {
	HandleMessage(context.Context,amqp.Delivery)
}

type HandlerFunc func(context.Context, amqp.Delivery)

func (f HandlerFunc) HandleMessage(ctx context.Context, m amqp.Delivery) {
	f(ctx, m)
}

type KeyHandlerFunc func(context.Context, amqp.Delivery) error









