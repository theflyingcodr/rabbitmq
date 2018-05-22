package events

import (
	"context"
)

type ConsumerConfig struct{
	Name string
	Durable bool
	Ttl int64
}

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
	Init(cfg ConsumerConfig) error
	// Prefix defines a common prefix to be added
	// to all queue names
	Prefix() string
	// Middleware can be used to implement custom
	// middleware which gets called before messages
	// are passed to handlers
	Middleware(next HandlerFunc) HandlerFunc
	// Setup is used to define the queues, keys and handlers
	// Config is passed which can be used to set QOS and consumer name
	Setup(ctx context.Context) (map[string]*ConsumerRoutes)
}

type ConsumerRoutes struct{
	Keys []string
	Handler HandlerFunc
}

