package consumer

import (
	"github.com/streadway/amqp"
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	"errors"
	"runtime/debug"
	"time"
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
	// AddBroker will register an exchange and n consumers
	// which will consume from that exchange
	AddBroker(context.Context, *ExchangeConfig, []Consumer) error
	// Start will setup all queues and routing keys
	// assigned to each consumer and then in turn start them
	Run(context.Context) (err error)
	// Middleware can be used to implement custom
	// middleware which gets called before messages
	// are passed to handlers
	Middleware(...HostMiddleware)
	// Stop can be called when you wish to shut down the host
	Stop(context.Context) error
}

type RabbitHost struct{
	c *HostConfig
	connection *amqp.Connection
	exchanges []Exchange
	channels map[string]*amqp.Channel
	middleware MiddlewareList
	connectionClose chan *amqp.Error
}

type Exchange struct{
	exchange *ExchangeConfig
	consumers []Consumer
}

// Init sets up the initial connection & quality of service
// to be used by all registered consumers
func NewConsumerHost(cfg *HostConfig) Host{
	host := &RabbitHost{
		exchanges:make([]Exchange, 0),
		channels:make(map[string]*amqp.Channel),
		c: cfg,
		connectionClose:make(chan *amqp.Error),
	}
	go host.rabbitConnector()
	host.connectionClose <- amqp.ErrClosed
	return host
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
	for h.connection == nil{}
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
		b.exchange.BuildExchange(ch)

		for _, c := range b.consumers {
			cfg, err := c.Init()
			if err != nil{
				log.Error(err)
				return err
			}
			if cfg == nil{
				cfg = &ConsumerConfig{}
			}
			queueChannel, err := h.connection.Channel()
			if err != nil{
				log.Error(err)
				return err
			}
			h.channels[cfg.GetName()] = queueChannel

			cfg.BuildConsumer(c, queueChannel, n, h.middleware)
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
	h.middleware = append(h.middleware, fn...)
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

// panicHandler intercepts panics from a consumer, logs
// the error and stack trace then nacks the message
func panicHandler(h HandlerFunc) HandlerFunc{
	return func(ctx context.Context, d amqp.Delivery) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("unknown error")
				}
				log.Errorf("panic handler recovered from unexpected panic, error: %s", err)
				log.Debugf("stack: %s", debug.Stack())
				d.Nack(false, false)
			}
		}()

		h(ctx, d)

	}
}

// errorHandler performs two functions
// it handles an error and returns an Ack if nil or a
// nack if err is not nil
// It also converts a KeyHandlerFunc to a HandlerFunc
// so middleware can be chained
func errorHandler(h KeyHandlerFunc) HandlerFunc{
	return func(ctx context.Context, d amqp.Delivery){
		err := h(ctx, d)
		if err != nil {
			log.Infof("error sending message with key %s and correlationid %v. Error: %s", d.RoutingKey, d.CorrelationId, err.Error())
			d.Nack(false, false)
		} else{
			d.Ack(false)
		}
	}
}

// HostMiddleware is
type HostMiddleware func(handler HandlerFunc) HandlerFunc

type MiddlewareList []HostMiddleware

func (h *RabbitHost) connect(){
	for {
		conn, err := amqp.Dial(h.c.Address)

		if err == nil {
			h.connection = conn
			break
		}

		log.Error(err)
		log.Infof("Trying to reconnect to RabbitMQ at %s\n", h.c.Address)
		time.Sleep(200 * time.Millisecond)
	}
}

func (h *RabbitHost) rabbitConnector() {
	println("hit connecter")
	var rabbitErr *amqp.Error

	for {
		rabbitErr = <-h.connectionClose
		if rabbitErr != nil {
			log.Infof("connecting to %s\n", h.c.Address)

			h.connect()
			h.connectionClose = make(chan *amqp.Error)
			for h.connection == nil {}
			h.connection.NotifyClose(h.connectionClose)
		}
	}
}