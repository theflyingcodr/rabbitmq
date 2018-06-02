# RabbitMq
A small wrapper library for [streadway/amqp](https://github.com/streadway/amqp) that provides an opinionated way of setting up RabbitMq consumers.

Don't worry about loosing connection to RabbitMq, the library will manage the connection and gracefully handle network issues and restart as soon as it has reached the Rabbit server.

Deadlettering is a first class citizen and all queues by default have their own dead letter queues setup, with no configuration required from you.

Sane defaults are set for Exchanges & Queues so the only required configuration from you is an amqp url and exchange names. Of course, you can override defaults by supplying your own configuration values for Exchanges & Queues.

Setting up Queues with Routing Keys is as easy as setting up an Http server with full support for Handlers & Middleware using familar patterns.

Queue setup borrows ideas fairly heavily from (Gizmo)[https://github.com/NYTimes/gizmo]

## A Consumer Example

Below is an example Consumer Implementation, yes, there is a simple Interface you can implement so you can be confident of compatibility.

``` go
// MyConsumer is an example consumer
type MyConsumer struct{
   // consumer.ConsumerConfig can be embedded here to override the defaults
}

// NewMyConsumer sets up and returns the consumer
// You could pass a customer consumer.ConsumerConfig here and return
// in the init method
func NewMyConsumer() *MyConsumer {
   return &MyConsumer{}
}

// Init is generally to be used to return a consumer.ConsumerConfig
// which has been passed to the consumer.
func(c *MyConsumer) Init() (*consumer.ConsumerConfig, error) {
   return nil, nil
}

// Prefix gets prepended to all queues registered
// to this consumer. An empty string can
// be returned which will mean no prefix will appear
func(c *MyConsumer) Prefix() string{
   return "azert"
}

// Middleware is to be used to add consumer level middleware, this will
// affect all queues registered & could be used to reject messages
// based on headers on body type for example
func(c *MyConsumer) Middleware(next consumer.HandlerFunc) consumer.HandlerFunc{
   return next
}

// Queues is used to setup Queues with routing keys and handlers
func(c *MyConsumer)Queues(ctx context.Context) (map[string]*consumer.Routes){
   return map[string]*consumer.Routes{
      "error":{ // queue name
		  Keys: []string{"test.error", "azert.error.#"}, // routing keys
		  DeliveryFunc:c.ErrorHandler, // handler func
	  },
	  "success":{ // queue name
		  Keys: []string{"test.success"}, // routing keys
		  DeliveryFunc:c.SuccessHandler, // handler func
		},
	}
}

// TestHandler is a really basic handler that returns an error
func (c *MyConsumer) ErrorHandler(ctx context.Context, m amqp.Delivery) error{
	logrus.Info("hit error handler")

	 // returning an error will automatically trigger a nack
	 // and send the message to the deadletter queue
	 return fmt.Errorf("oh no i failed")
}

// SuccessHandler is a really basic handlerthat returns a success
func (c *MyConsumer) SuccessHandler(ctx context.Context, m amqp.Delivery) error{
   logrus.Info("hit success handler")

   // returning nil will automatically trigger an ack
   return nil
}
```
## Running it
To run the above consumer example we add a main.go file with the below code

``` go
func main(){
   logrus.SetLevel(logrus.DebugLevel)

   // setup & configure our Host
   host := consumer.NewConsumerHost(&consumer.HostConfig{Address:"amqp://guest:guest@localhost:5672/"})

   // this is a basic example, config would usually be setup via env vars or a file
   exchange := &consumer.ExchangeConfig{Name:"test"}

   // to add global custom middleware, just chain it here
   // if you don't need global middle just remove this method
   host.Middleware(consumer.MessageDump, consumer.JsonHandler)

   // this registers a list of consumers with an exchange which
   // is defined in configuration // multiple brokers can be added if required
   host.AddBroker(context.Background(), exchange, []consumer.Consumer{NewMyConsumer()})

   // run the consumer, it will exit on a setup error
   // or on a graceful shutdown ie ctrl+c
   if err := host.Run(context.Background()); err != nil{
      logrus.Error(err)
   }
}
```
##  Runable Example
An example implementation can be found under the *examples* folder. ``` go run main.go``` will kick it off and you can try publishing messages and observe the results.

For testing I recommend using docker with the rabbitmq:management-alpine images using the below command

```bash
docker run -d --hostname test-rabbit --name rabbitmq-test -p 5672:5672 -p 15672:15672 rabbitmq:management-alpine
```

## Publishing
A publisher implementation will appear soon, for testing either write your own or using the RabbitMq ui for now.

## Middleware
A key feature of Go & especially http servers is the ability to write & chain middleware.

This library fully supports middleware and allows the user to define and include their own. The interface is

```go
type MessageHandler interface {
   HandleMessage(context.Context,amqp.Delivery)
}
```
And the HandlerFunc is
```go
type HandlerFunc func(context.Context, amqp.Delivery)
```
There are currently two bits of middleware that can be used by users, they are *JsonHandler* & *MessageDump*, you can look at them in the *consumer/middleware.go* file for examples of writing middleware.

Middleware can be added by users at two levels
### Consumer Level Middleware

``` go
func(c *MyConsumer) Middleware(next consumer.HandlerFunc) consumer.HandlerFunc{
   return consumer.JsonHandler(consumer.MessageDump(next))
}
```
This shows the pre written middleware being added to the Consumers Middleware method.

```go
func(c *MyConsumer) Middleware(next consumer.HandlerFunc) consumer.HandlerFunc{
   return func(ctx context.Context, d amqp.Delivery) {
      if d.ContentType != "mycompany/customtype"{
         d.Nack(false, false)
         return
  }
      next(ctx, d)
   }
}
```
Above shows a bit of custom middleware written by a user that rejects messages based on contentType

### Global Middleware
You can also add middleware that gets executed for **every** consumer registered to the host. This is done as shown in your main.go file

```go
// setup & configure our Host
host := consumer.NewConsumerHost(&consumer.HostConfig{Address:"amqp://guest:guest@localhost:5672/"})

// this is a basic example, config would usually be setup via env vars or a file
exchange := &consumer.ExchangeConfig{Name:"test"}

// to add global custom middleware, just chain it here
// if you don't need global middle just remove this method
host.Middleware(consumer.MessageDump, consumer.JsonHandler)
```
Note the final line, this is a method that can be used to list a chain of middleware. The method itself is optional, if you don't want to add global middleware, don't add the method, simple.

Here i am using the MessageDump debug middleware & the JsonHandler, this will reject messages that aren't "application/json"

As with the consumer, you can also define and chain your own middleware. To keep things neat, it's best to define these funcs elsewhere and then add them to the chain as shown above.