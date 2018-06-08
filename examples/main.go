package main

import (
	"github.com/sirupsen/logrus"
	"context"
	"github.com/azert-software/rabbitmq/consumer"
	"github.com/streadway/amqp"
	"fmt"
)

func main(){
	logrus.SetLevel(logrus.DebugLevel)

	ec := consumer.NewErrorChannel()
	// setup & configure our Host
	host := consumer.NewConsumerHost(&consumer.HostConfig{Address:"amqp://guest:guest@localhost:5672/"}, ec)

	// this is a basic example, config would usually be setup via env vars or a file
	exchange := &consumer.ExchangeConfig{Name:"test"}

	// to add global custom middleware, just chain it here
	// if you don't need global middle just remove this method
	host.Middleware(consumer.MessageDump, consumer.JsonHandler)

	// this registers a list of consumers with an exchange which
	// is defined in configuration
	// multiple brokers can be added if required
	host.AddBroker(context.Background(), exchange, []consumer.Consumer{NewMyConsumer()})

	go func(){
		for {
			hc := consumer.NewHealthCheckServer(&consumer.HealthCheckConfig{
				Port:9999,
			},host.GetConnectionStatus, ec)
			hc.SetupRoutes()
			hc.Run()
		}
	}()
	// run the consumer, it will exit on a setup error
	// or on a graceful shutdown ie ctrl+c
	if err := host.Run(context.Background()); err != nil{
		logrus.Error(err)
	}
}

// MyConsumer is an example consumer
type MyConsumer struct{
	// consumer.ConsumerConfig can be embedded here to override the defaults
}

// NewMyConsumer sets up and returns the consumer
// You could pass a customer consumer.ConsumerConfig here and return
// in the init method
func NewMyConsumer() *MyConsumer {
	return &MyConsumer{}
}

// Init is generally to be used to return a consumer.ConsumerConfig
// which has been passed to the consumer
func(c *MyConsumer) Init() (*consumer.ConsumerConfig, error) {
	return nil, nil
}

// Prefix gets prepended to all queues registered
// to this consumer
func(c *MyConsumer) Prefix() string{
	return "azert"
}

// Middleware is to be used to add consumer level middleware, this will
// affect all queues registered & could be used to reject messages
// based on headers on body type for example
func(c *MyConsumer) Middleware(next consumer.HandlerFunc) consumer.HandlerFunc{
	return func(ctx context.Context, d amqp.Delivery) {
		if d.ContentType != "mycompany/customtype"{
			d.Nack(false, false)
			return
		}

		next(ctx, d)
	}

}

// Queues is used to setup Queues with routing keys and handlers
func(c *MyConsumer)Queues(ctx context.Context) (map[string]*consumer.Routes){
	return map[string]*consumer.Routes{
		"error":{ // queue name
			Keys: []string{"test.error", "azert.error.#"}, // routing keys
			DeliveryFunc:c.ErrorHandler, // handler func
		},
		"success":{ // queue name
			Keys: []string{"test.success", "azert.#"}, // routing keys
			DeliveryFunc:c.SuccessHandler, // handler func
		},
	}
}

// TestHandler is a really basic example that returns an error
func (c *MyConsumer) ErrorHandler(ctx context.Context, m amqp.Delivery) error{
	logrus.Info("hit error handler")

	// returning an error will automatically trigger a nack
	// and send the message to the deadletter queue
	return fmt.Errorf("oh no i failed")
}

// SuccessHandler is a really basic example that returns a success
func (c *MyConsumer) SuccessHandler(ctx context.Context, m amqp.Delivery) error{
	logrus.Info("hit success handler")

	// returning nil will automatically trigger an ack
	return nil
}