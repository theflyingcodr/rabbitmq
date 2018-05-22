package rabbitmq

import (
	"bitbucket.org/azertsoftware/events"
	"github.com/streadway/amqp"
)



type RabbitHost struct{
	c *events.HostConfig
	connection *amqp.Connection
}

func NewRabbitHost() events.Host{
	return &RabbitHost{}
}

func (h *RabbitHost) Init(cfg *events.HostConfig) (err error){
	h.c = cfg
	h.connection, err = amqp.Dial(h.c.Address)
	return
}

func (h *RabbitHost) Start(brokers []events.Broker) error{
	for i, b := range
}

func (h *RabbitHost) Stop() error{

}

type BrokerSetup struct{
	Exchange ExchangeConfig
	Consumers []events.Consumer
}