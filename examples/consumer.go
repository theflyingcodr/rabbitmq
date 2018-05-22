package main

import (
	"bitbucket.org/azertsoftware/events"
	"context"
)

func main(){

}

type MyConsumer struct{

}

func NewMyConsumer() events.Consumer {
	return &MyConsumer{}
}

func (c *MyConsumer) Prefix() string{
	return ""
}
func (c *MyConsumer) Middleware(h events.HandlerFunc) events.HandlerFunc{
	return func(m events.BasicMessage) error {
		return h(m)
	}
}

func (c *MyConsumer) Setup(ctx context.Context) (map[string]*events.ConsumerRoutes, ){
	return map[string]*events.ConsumerRoutes{"mark.queue":{
			Keys: []string{"test.message", "mark.#"},
			Handler:c.TestHandler,
		},
	}
}

func (c *MyConsumer) TestHandler(m events.BasicMessage) error{
	return nil
}