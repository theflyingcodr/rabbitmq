package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/azert-software/amqp"
	log "github.com/Sirupsen/logrus"
	"context"
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
	h.brokers = append(h.brokers, Broker{exchange:cfg, consumers:consumers})

	return nil
}

// Start will setup all queues and routing keys
// assigned to each consumer and then in turn start them
func (h *RabbitHost) Run() (err error){
	ch, err := h.connection.Channel()
	if err != nil{
		return
	}
	for _, b := range h.brokers{
		n, err := b.exchange.GetName()
		if err != nil{
			return
		}
		h.BuildExchange(ch, b)
		for _, c := range b.consumers {
			cfg, err := c.Init()
			if err != nil{
				return
			}

			for k, r := range c.Setup(context.Background()){
				go func(){
					// each consumer has its own channel & each queue has its own consumer
					queueChannel, err := h.connection.Channel()
					if err != nil{
						return
					}

					if err = queueChannel.Qos(int(cfg.GetPrefetchCount()), int(cfg.GetPrefetchSize()), false); err != nil{
						return
					}
					_, err = queueChannel.QueueDeclare(cfg.GetName(), cfg.GetDurable(), cfg.GetAutoDelete(), cfg.GetExclusive(), cfg.GetNoWait(), cfg.Args)
					if err != nil{
						return
					}
					for _, key := range r.Keys{
						if err = queueChannel.QueueBind(k, key, n, cfg.GetNoWait(), cfg.Args ); err != nil{
							return
						}


					}
				}()

			}

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
		log.Errorf("error when setting up exchange %s: %s",n, err.Error())
		return
	}
	log.Debugf("%s exchange setup success", n)
	return
}