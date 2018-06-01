package main

import (
	"github.com/sirupsen/logrus"
	"context"
	"github.com/azert-software/rabbitmq/consumer"
	"github.com/streadway/amqp"
	"fmt"
)

func main(){
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	host := consumer.RabbitHost{}
	if err := host.Init( context.Background(), &consumer.HostConfig{Address:"amqp://guest:guest@localhost:5672/",}); err != nil{
		return
	}
	eCfg := &consumer.ExchangeConfig{
		Name:"test",
	}

	host.Middleware( consumer.MessageDump, consumer.JsonHandler)

	host.AddBroker(context.Background(),eCfg,[]consumer.Consumer{NewMyConsumer()})
	if err := host.Run(context.Background()); err != nil{
		logrus.Error(err)
	}
}

type config struct{
	ConsumerHost consumer.HostConfig
}

type MyConsumer struct{

}

func NewMyConsumer() *MyConsumer {
	return &MyConsumer{}
}

func(c *MyConsumer) Init() (*consumer.ConsumerConfig, error) {
	return nil, nil
}

func(c *MyConsumer) Prefix() string{
	return ""
}

func(c *MyConsumer) Middleware(next consumer.HandlerFunc) consumer.HandlerFunc{
	return next
}

func(c *MyConsumer)Queues(ctx context.Context) (map[string]*consumer.Routes){
	return map[string]*consumer.Routes{
		"mark.rabbitqueue":{
			Keys: []string{"test.message", "mark.#"},
			DeliveryFunc:c.TestHandler,
		},
	}
}

func (c *MyConsumer) TestHandler(ctx context.Context, m amqp.Delivery) error{
	println("hit handler")
	return fmt.Errorf("oh no i failed")
}