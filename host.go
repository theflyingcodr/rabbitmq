package messaging

type HostConfig struct{
	Address string
	Qos int
}

func (h *HostConfig) GetQos() int{
	if h.Qos == 0{
		return 10
	}

	return h.Qos
}

// Host is the container which is used
// to host all consumers that are registered.
// It is responsible for the amqp connection
// starting & gracefully stopping all running consumers
// h := NewRabbitHost().Init(cfg.Host)
// h.AddBroker(NewBroker(cfg.Exchange, [])
type Host interface{
	// Init sets up the initial connection & quality of service
	// to be used by all registered consumers
	Init(*HostConfig) (err error)
	// AddBroker will register an exchange and n consumers
	// which will consume from that exchange
	AddBroker(*ExchangeConfig, []Consumer) error
	// Start will setup all queues and routing keys
	// assigned to each consumer and then in turn start them
	Run() (err error)
}

// BasicMessage is a generic message that
// can get transformed via middleware
// to something more useful
type BasicMessage struct{
	ID int64
	Headers map[string]string
	Body interface{}
}

type HandlerFunc func(BasicMessage) error

func (f HandlerFunc) HandleMessage(m BasicMessage) error{
	return f(m)
}

