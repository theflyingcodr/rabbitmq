package messaging

import "github.com/labstack/gommon/log"

func Logger(h HandlerFunc) HandlerFunc{
	return func(msg AMQPMessage) error{
		log.Infof("message with ID %v received", msg.ID)

		return h(msg)
	}
}
