package messaging

import (
	"log"
	"context"
)

func Logger(h HandlerFunc) HandlerFunc{
	return func(ctx context.Context, msg AMQPMessage) error{
		log.Printf("message with ID %v received", msg.ID)

		e:= h(ctx, msg)
		return e
	}
}
