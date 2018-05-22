package rabbitmq

import (
	"github.com/streadway/amqp"
	"errors"
)



func (e *ExchangeConfig) BuildExchange(c *amqp.Channel) (err error){
	eName, err := e.GetName()
	if err != nil{
		return
	}
	err = c.ExchangeDeclare(eName, e.GetType(), e.GetDurable(), e.GetAutoDelete(), e.GetInternal(), false, e.GetArgs())
	return
}