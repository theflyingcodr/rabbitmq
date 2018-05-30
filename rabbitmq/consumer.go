package rabbitmq

import (
	"context"
	"github.com/azert-software/messaging"
)

type RabbitConsumer interface{
	// Init can be used to get a custom
	// consumer config. If it returns nil a consumer
	// with default params will be setup
	Init() (*messaging.ConsumerConfig, error)
	// Prefix defines a common prefix to be added
	// to all queue names
	Prefix() string
	Middleware(messaging.HandlerFunc) messaging.HandlerFunc
	DeliveryMiddleware(DeliveryHandlerFunc) DeliveryHandlerFunc
	Queues(ctx context.Context) (map[string]*DeliveryHandlerFunc)

}

type DeliveryRoutes struct{
	Keys []string
	DeliveryFunc DeliveryHandlerFunc
}


