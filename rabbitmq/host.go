package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/azert-software/amqp"
	"fmt"
	"github.com/teamwork/log"
)



type RabbitHost struct{
	c *messaging.HostConfig
	connection *amqp.Connection
	brokers []Broker
}

type Broker struct{
	exchange *messaging.ExchangeConfig
	consumers []messaging.Consumer
}


// Init sets up the initial connection & quality of service
// to be used by all registered consumers
func (h *RabbitHost) Init(cfg *messaging.HostConfig) (err error){
	h.brokers = make([]Broker, 0)
	h.c = cfg
	h.connection, err = amqp.Dial(h.c.Address)
	return
}

// AddBroker will register an exchange and n consumers
// which will consume from that exchange
func (h *RabbitHost) AddBroker(cfg *messaging.ExchangeConfig, consumers []messaging.Consumer) error {
	n, err := cfg.GetName()
	if err != nil{
		return err
	}
	h.brokers = append(h.brokers, Broker{exchange:cfg, consumers:consumers})

	return nil
}

// Start will setup all queues and routing keys
// assigned to each consumer and then in turn start them
func (h *RabbitHost) Run() (err error){
	ch, err := h.connection.Channel()
	if err != nil{
		return err
	}
	for _, b := range h.brokers{
		h.BuildExchange(ch, b)
		for _, c := range b.consumers {

		}
	}

	return ch.Close()
}

// BuildExchange builds an exchange
func (h *RabbitHost) BuildExchange(ch *amqp.Channel, b Broker) (err error){
	ex := b.exchange
	n, err := ex.GetName()
	if err != nil{
		return err
	}
	log.Debugf("setting up %s exchange", n)
	if err = ch.ExchangeDeclare(n, ex.GetType(), ex.GetDurable(), ex.GetAutoDelete(), ex.GetInternal(), false, ex.GetArgs()); err != nil{
		log.Errorf(err, "error when setting up exhange %s",n)
		return
	}

	log.Debugf("%s exchange setup success", n)
	return
}