package consumer

import (
	"context"
	"github.com/pborman/uuid"
	"fmt"
)

// Consumer is an interface which can be implemented
// to create a consumer
//
// The consumer can setup multiple queues, define
// n routing keys for each queue and in turn assign
// a handler to manage the messages received
//
// Custom middleware can be added using the Middleware
// method
//
// Example implementation shown below:
//
// 	type MyConsumer struct{
//
//	}
//
//	func NewMyConsumer() events.Consumer {
//		return &MyConsumer{}
//	}
//
//	func (c *MyConsumer) Prefix() string{
//		return ""
//	}
//
//	func (c *MyConsumer) Middleware(h events.HandlerFunc) events.HandlerFunc{
//		return func(m events.BasicMessage) error {
//			return h(m)
//		}
//	}
//
//	func (c *MyConsumer) Setup(ctx context.Context) (map[string]*events.ConsumerRoutes, ){
//		return map[string]*events.ConsumerRoutes{"mark.queue":{
//				Keys: []string{"test.message", "mark.#"},
//				Handler:c.TestHandler,
//			},
//		}
//	}
//
//	func (c *MyConsumer) TestHandler(m events.BasicMessage) error{
//		return nil
//	}
//
type Consumer interface{
	// Init can be used to get a custom
	// consumer config. If it returns nil a consumer
	// with default params will be setup
	Init() (*ConsumerConfig, error)
	// Prefix defines a common prefix to be added
	// to all queue names
	Prefix() string
	// Middleware can be used to implement custom
	// middleware which gets called before messages
	// are passed to handlers
	Middleware(HandlerFunc) HandlerFunc
	// Queues is used to define the queues, keys and handlers
	// Config is passed which can be used to set QOS and consumer name
	Queues(context.Context) (map[string]*Routes)
}

// Routes contains a set of routing keys and
// a handlerFunc that will be used to process
// messages meeting the routing keys
type Routes struct{
	Keys []string
	DeliveryFunc KeyHandlerFunc
}

// ConsumerConfig defines the setup of a consumer
// If this isn't set default values will be used.
// To set a custom config for a consumer setup a new
// consumer struct, pass the config and return it in the Init() method
type ConsumerConfig struct{
	Name string
	Durable *bool
	AutoDelete *bool
	NoWait *bool
	Exclusive *bool
	Ttl *uint
	PrefetchCount *uint
	PrefetchSize *uint
	Args map[string]interface{}
}

// GetName returns the consumer name if set in config
// otherwise it returns a random uuid
func(c *ConsumerConfig) GetName() string{
	if c.Name == ""{
		return uuid.NewUUID().String()
	}

	return fmt.Sprintf("%s-%s",c.Name, uuid.NewUUID().String())
}

// GetDurable returns the type of durability
// set in config, if nil then it returns a
// default of true
func(c *ConsumerConfig) GetDurable() bool{
	if c.Durable == nil{
		return true
	}
	return *c.Durable
}

// GetPrefetchCount returns the Qos value for number
// of messages pulled from the queue at a time
// default is 0 which will pull the default count for most libs
func(c *ConsumerConfig) GetPrefetchCount() uint{
	if c.PrefetchCount == nil{
		return 0
	}

	return *c.PrefetchCount
}

func(c *ConsumerConfig) GetPrefetchSize() uint{
	if c.PrefetchSize == nil{
		return 0
	}

	return *c.PrefetchSize
}

// GetAutoDelete determines whether the queue is deleted
// on server restart, default is false
func(c *ConsumerConfig) GetAutoDelete() bool{
	if c.AutoDelete == nil{
		return false
	}
	return *c.AutoDelete
}

// GetNoWait When true, the queue will assume to be declared on the server.  A
//channel exception will arrive if the conditions are met for existing queues
//or attempting to modify an existing queue from a different connection.
// default is false
func(c *ConsumerConfig) GetNoWait() bool{
	if c.NoWait == nil{
		return false
	}
	return *c.NoWait
}

// GetExclusive queues are only accessible by the connection that declares them and
//will be deleted when the connection closes.
// default is false
func(c *ConsumerConfig) GetExclusive() bool{
	if c.Exclusive == nil{
		return false
	}
	return *c.Exclusive
}

// GetArgs gets a table of arbitrary arguments
// which are passed to the exchange
func (e *ConsumerConfig) GetArgs() map[string]interface{}{
	if e.Args == nil{
		return make(map[string]interface{})
	}
	return e.Args
}


