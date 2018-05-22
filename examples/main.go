package main

import (
	"bitbucket.org/azertsoftware/events/rabbitmq"
	"bitbucket.org/azertsoftware/events"
	"github.com/derekparker/delve/service"
)

func main(){
	svcBroker := NewBroker(service.Config{}, []{NewMyConsumer(myConfig)} )
	h := host.Init(cfg)


	rabbitmq.NewRabbitHost(&events.HostConfig{}).
		Init()
}

type config struct{
	ConsumerHost events.HostConfig
	MyConsumer Rabbit
}

type Rabbit struct{
	Exchange rabbitmq.ExchangeConfig
	Consumer events.ConsumerConfig
}
