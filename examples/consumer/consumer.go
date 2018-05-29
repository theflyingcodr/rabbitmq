package consumer

import (
	"context"
	"github.com/azert-software/messaging"
)

type MyConsumer struct{

}

func NewMyConsumer() *MyConsumer {
	return &MyConsumer{}
}

func(c *MyConsumer) Init() (*messaging.ConsumerConfig, error) {
	return nil, nil
}

func(c *MyConsumer) Prefix() string{
	return ""
}

func(c *MyConsumer) Middleware(next messaging.HandlerFunc) messaging.HandlerFunc{
	return nil
}

func(c *MyConsumer) Setup(ctx context.Context) (map[string]*messaging.ConsumerRoutes){
	return map[string]*messaging.ConsumerRoutes{"mark.queue":{
			Keys: []string{"test.message", "mark.#"},
			Handler:c.TestHandler,
		},
	}
}

func (c *MyConsumer) TestHandler(m messaging.BasicMessage) error{
	return nil
}