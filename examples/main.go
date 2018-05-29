package main

import (
	"github.com/azert-software/messaging"
	"github.com/azert-software/messaging/rabbitmq"
	"github.com/azert-software/messaging/examples/consumer"
	"github.com/sirupsen/logrus"
	"context"
)

func main(){
	host := rabbitmq.RabbitHost{}
	if err := host.Init( context.Background(), &messaging.HostConfig{Address:"amqp://guest:guest@localhost:5672/",}); err != nil{
		return
	}
	eCfg := &messaging.ExchangeConfig{
		Name:"test",
	}
	host.AddBroker(context.Background(),eCfg,[]messaging.Consumer{consumer.NewMyConsumer()})
	if err := host.Run(context.Background()); err != nil{
		logrus.Error(err)
	}
}

type config struct{
	ConsumerHost messaging.HostConfig
}
