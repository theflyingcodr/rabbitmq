package consumer

import (
	"github.com/streadway/amqp"
	"context"
	log "github.com/sirupsen/logrus"
	"fmt"
)

func JsonHandler(h HandlerFunc) HandlerFunc{
	return func(ctx context.Context, d amqp.Delivery) {
		if d.ContentType != "application/json"{
			log.Infof("message received with invalid content-type %s, expected application/json", d.ContentType)
			d.Nack(false, false)
			return
		}
		h(ctx, d)
	}
}

// MessageDump will output the entire amqp message
// with the body converted to a string
// Handy for debugging
func MessageDump(h HandlerFunc) HandlerFunc{
	return func(ctx context.Context, d amqp.Delivery) {
		log.WithFields(log.Fields{
			"fullmessage":fmt.Sprintf("%+v", d),
			"body":fmt.Sprintf("%s", d.Body),
		}).Info("new message received")
		h(ctx, d)
	}
}
