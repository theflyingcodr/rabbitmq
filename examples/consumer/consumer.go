package consumer

import (
	"context"
	"github.com/azert-software/messaging"
	"github.com/streadway/amqp"
	"github.com/azert-software/messaging/rabbitmq"
)

type MyConsumer struct{

}

func NewMyConsumer() *MyConsumer {
	return &MyConsumer{}
}

func(c *MyConsumer) Init() (*messaging.ConsumerConfig, error) {
	return nil, nil
}

func(c *MyConsumer) Prefix() string{
	return ""
}

func(c *MyConsumer) Middleware(next messaging.HandlerFunc) messaging.HandlerFunc{
	return nil
}

func(c *MyConsumer) Setup(ctx context.Context) (map[string]*messaging.ConsumerRoutes){
	return map[string]*messaging.ConsumerRoutes{"mark.queue":{
			Keys: []string{"test.message", "mark.#"},
			Handler:c.TestHandler,
		},
	}
}

func(c *MyConsumer)DeliveryMiddleware(rabbitmq.DeliveryHandlerFunc) rabbitmq.DeliveryHandlerFunc{
	return func(ctx context.Context, delivery amqp.Delivery) error {
		delivery.Ack(false)
		return nil
	}
}

func(c *MyConsumer)SetupDelivery(ctx context.Context) (map[string]*rabbitmq.DeliveryRoutes){
	return map[string]*rabbitmq.DeliveryRoutes{
		"mark.rabbitqueue":{
			Keys: []string{"test.message", "mark.#"},
			DeliveryFunc:c.HandleDelivery,
		},
	}
}

func (c *MyConsumer) TestHandler(ctx context.Context, m messaging.AMQPMessage) error{
	return nil
}

func (h *MyConsumer) HandleDelivery(ctx context.Context, delivery amqp.Delivery) error{
	return nil
}