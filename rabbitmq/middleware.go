package rabbitmq

import (
	"github.com/azert-software/messaging"
	"github.com/streadway/amqp"
	"context"
)

func DeliveryToMessage(h DeliveryHandlerFunc) (handlerFunc messaging.HandlerFunc){
}

type DeliveryHandlerFunc func(ctx context.Context, delivery amqp.Delivery) error

func(h DeliveryHandlerFunc) HandleMessage(ctx context.Context, m messaging.AMQPMessage) error {
	e := h(ctx, amqp.Delivery{})
	return e
}

func (h *DeliveryHandlerFunc) HandleDelivery(ctx context.Context, delivery amqp.Delivery) error{
	return nil
}

func LoggingHandler(ctx context.Context, delivery amqp.Delivery) error{

}
