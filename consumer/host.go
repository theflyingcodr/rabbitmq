package consumer

import (
	"github.com/streadway/amqp"
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"fmt"
)

// HostConfig contains global config
// used for the rabbit connection
type HostConfig struct{
	Address string
}

// Host is the container which is used
// to host all consumers that are registered.
// It is responsible for the amqp connection
// starting & gracefully stopping all running consumers
// h := NewRabbitHost().Init(cfg.Host)
// h.AddBroker(NewBroker(cfg.Exchange, [])
type Host interface{
	// Init sets up the initial connection & quality of service
	// to be used by all registered consumers
	Init(context.Context, *HostConfig) (err error)
	// AddBroker will register an exchange and n consumers
	// which will consume from that exchange
	AddBroker(context.Context, *ExchangeConfig, []Consumer) error
	// Start will setup all queues and routing keys
	// assigned to each consumer and then in turn start them
	Run(context.Context) (err error)
	// Middleware can be used to implement custom
	// middleware which gets called before messages
	// are passed to handlers
	Middleware(...MiddlewareList)
	// Stop can be called when you wish to shut down the host
	Stop(context.Context) error
}

type RabbitHost struct{
	c *HostConfig
	connection *amqp.Connection
	exchanges []Exchange
	channels map[string]*amqp.Channel
	middleware MiddlewareList
}

type Exchange struct{
	exchange *ExchangeConfig
	consumers []Consumer
}

// Init sets up the initial connection & quality of service
// to be used by all registered consumers
func (h *RabbitHost) Init(ctx context.Context, cfg *HostConfig) (err error){
	h.exchanges = make([]Exchange, 0)
	h.channels = make(map[string]*amqp.Channel)
	h.c = cfg
	h.connection, err = amqp.Dial(h.c.Address)
	if err != nil{
		log.Errorf("error dialing rabbit %v", err)
	}
	return
}

// AddBroker will register an exchange and n consumers
// which will consume from that exchange
func (h *RabbitHost) AddBroker(ctx context.Context, cfg *ExchangeConfig, consumers []Consumer) error {
	h.exchanges = append(h.exchanges, Exchange{exchange:cfg, consumers:consumers})

	return nil
}

// Start will setup all queues and routing keys
// assigned to each consumer and then in turn start them
func (h *RabbitHost) Run(ctx context.Context) (err error){
	ch, err := h.connection.Channel()
	if err != nil{
		log.Errorf("error when getting channel from connection: %v", err.Error())
		return err
	}
	for _, b := range h.exchanges{
		n, err := b.exchange.GetName()
		if err != nil{
			log.Error(err)
			return err
		}
		h.BuildExchange(ch, b)
		for _, c := range b.consumers {
			cfg, err := c.Init()
			if err != nil{
				log.Error(err)
				return err
			}
			if cfg == nil{
				cfg = &ConsumerConfig{}
			}

			for k, r := range c.Queues(context.Background()){
				go func() {
					log.Infof("setting up queue %s", k)
					// each consumer has its own channel & each queue has its own consumer
					queueChannel, err := h.connection.Channel()
					if err != nil {
						log.Fatal(err)
					}
					defer queueChannel.Close()
					h.channels[cfg.GetName()] = queueChannel

					if err = queueChannel.Qos(int(cfg.GetPrefetchCount()), int(cfg.GetPrefetchSize()), false); err != nil {
						log.Fatal(err)
					}

					dlx := fmt.Sprintf("%s.deadletter", n)
					a := cfg.GetArgs()
					a["x-dead-letter-exchange"] = dlx
					_, err = queueChannel.QueueDeclare(k, cfg.GetDurable(), cfg.GetAutoDelete(), cfg.GetExclusive(), cfg.GetNoWait(), a)
					if err != nil {
						log.Fatal(err)
					}
					dlq := fmt.Sprintf("%s.deadletter",k)
					_, err = queueChannel.QueueDeclare(dlq, true, false, false, false, nil)
					if err != nil {
						log.Fatal(err)
					}

					for _, key := range r.Keys {
						log.Debugf("binding key %s to queue %s",key, k)
						if err = queueChannel.QueueBind(k, key, n, cfg.GetNoWait(), cfg.Args); err != nil {
							log.Fatal(err)
						}
						if err = queueChannel.QueueBind(dlq, key, fmt.Sprintf("%s.deadletter", n), false, nil); err != nil {
							log.Fatal(err)
						}
					}
					log.Infof("queue %s setup",k)

					msgs, err := queueChannel.Consume(k,cfg.Name,false,cfg.GetExclusive(),false,cfg.GetNoWait(), cfg.Args)
					if err != nil{
						log.Fatal(err)
					}

					var fn HandlerFunc
					ch, _ := h.connection.Channel()
					for i := range h.middleware.middleware {
						fn = h.middleware.middleware[len(h.middleware.middleware)-1-i](c.Middleware(errorHandler(ch, r.DeliveryFunc)))
					}

					for d := range msgs{
						panicHandler(fn).HandleMessage(context.Background(), d)
					}
				}()
			}
		}
	}

	ch.Close() // discard this channel
	log.Infof("host started")
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	return h.Stop( ctx)
}

func (h *RabbitHost) Middleware(fn ...HostMiddleware) {
	h.middleware = MiddlewareList{append(([]HostMiddleware)(nil), fn...)}
}

func (h *RabbitHost) Stop(context.Context) error{
	log.Infof("shutting down host")

	for k, v := range h.channels{
		log.Infof("closing channel for queue %s", k)
		if err := v.Close(); err != nil{
			log.Errorf("error when closing channel %s: %s",k, err)
			continue
		}
		log.Infof("channel for queue %s closed successfully", k)
	}

	return h.connection.Close()
}

// BuildExchange builds an exchange
func (h *RabbitHost) BuildExchange(ch *amqp.Channel, b Exchange) (err error){
	ex := b.exchange
	n, err := ex.GetName()
	if err != nil{
		log.Error(err)
		return err
	}

	log.Debugf("setting up %s exchange", n)

	dlx := fmt.Sprintf("%s.deadletter", n)
	if err = ch.ExchangeDeclare(dlx, ex.GetType(), ex.GetDurable(), ex.GetAutoDelete(), ex.GetInternal(), false, ex.GetArgs()); err != nil{
		log.Errorf("error when setting up deadletter exchange %s: %s", dlx)
		return
	}

	args := ex.GetArgs()
	args["x-dead-letter-exchange"] = dlx

	if err = ch.ExchangeDeclare(n, ex.GetType(), ex.GetDurable(), ex.GetAutoDelete(), ex.GetInternal(), false, ex.GetArgs()); err != nil{
		log.Errorf("error when setting up exchange %s: %s",n, err.Error())
		return
	}


	log.Debugf("%s exchange setup success", n)
	return
}