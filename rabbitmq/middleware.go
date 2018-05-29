package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/azert-software/messaging"
)

func DeliveryToMessage(delivery amqp.Delivery) (msg messaging.AMQPMessage){
		msg = messaging.AMQPMessage{
			ID:              delivery.MessageId,
			Body:            delivery.Body,
			Headers:         delivery.Headers,
			ContentEncoding: delivery.ContentEncoding,
			ContentType:     delivery.ContentType,
			CorrelationID:   delivery.CorrelationId,
			Created:         delivery.Timestamp,
			ReplyTo:         delivery.ReplyTo,
			Subject:         delivery.RoutingKey,
			UserID:          delivery.UserId,
			Expiry:          delivery.Expiration,
		}

		return
}
