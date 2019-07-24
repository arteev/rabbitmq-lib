package rabbit

import (
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

//channel RabbitMQ channel operations
type Channel interface {
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error

	Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	Close() error
}

type Connection interface {
	NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
	Close() error
	Channel() (Channel, error)
}

type RabbitConnection struct {
	mu     sync.RWMutex
	url    string
	closed bool
	done   chan struct{}
	Connection
}

type wrapConnection struct {
	*amqp.Connection
}

func (w *wrapConnection) Channel() (Channel, error) {
	return w.Connection.Channel()
}

func NewRabbitConnection(connection Connection, connectionString string, timeout int) (*RabbitConnection, error) {
	done := make(chan struct{}, 0)

	if connection == nil {
		conn, err := amqp.DialConfig(connectionString,
			amqp.Config{
				Dial: amqp.DefaultDial(time.Second * time.Duration(timeout)),
			})

		if err != nil {
			return nil, err
		}
		connection = &wrapConnection{conn}
	}
	newConnection := &RabbitConnection{
		done:       done,
		Connection: connection,
	}

	newConnection.waitClose()
	return newConnection, nil
}

func (f *RabbitConnection) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	close(f.done)
	return f.Connection.Close()
}

func (f *RabbitConnection) Channel() (Channel, error) {
	return f.Connection.Channel()
}

func (f *RabbitConnection) Connected() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return !f.closed
}

func (f *RabbitConnection) waitClose() {
	go func() {
		chError := make(chan *amqp.Error)
		f.NotifyClose(chError)
		for {
			select {
			case rabbitErr := <-chError:
				if rabbitErr != nil {
					log.Println("RabbitMQ disconnected ", rabbitErr)
				}
				f.mu.Lock()
				f.closed = true
				f.mu.Unlock()
				return
			case <-f.done:
				log.Println("exit rabbit reconnect")
				return
				break
			}
		}
	}()
}
