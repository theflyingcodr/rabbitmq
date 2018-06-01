package consumer

import (
	"github.com/streadway/amqp"
	"context"
	log "github.com/sirupsen/logrus"
	"errors"
	"runtime/debug"
	"fmt"
)

func loggingMiddleware(h HandlerFunc) HandlerFunc {
	return func(ctx context.Context, d amqp.Delivery){
		log.Infof("new message received %+v", d)

		h(ctx, d)
	}
}

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

func errorHandler(ch *amqp.Channel, h KeyHandlerFunc) HandlerFunc{
	return func(ctx context.Context, d amqp.Delivery){
		err := h(ctx, d)
		if err != nil {
			log.Infof("error sending message with key %s & id %v. Error: %s", d.RoutingKey, d.MessageId, err.Error())
			d.Nack(false, false)
		} else{
			d.Ack(false)
		}
	}
}

func JsonHandler(h HandlerFunc) HandlerFunc{
	return func(ctx context.Context, d amqp.Delivery) {
		fmt.Println("hit")
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

type HostMiddleware func(handler HandlerFunc) HandlerFunc

type MiddlewareList struct{
	middleware []HostMiddleware
}

