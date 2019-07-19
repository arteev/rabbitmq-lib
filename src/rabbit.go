package main

import (
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

type RabbitConnection struct {
	mu     sync.RWMutex
	url    string
	closed bool
	done   chan struct{}
	*amqp.Connection
}

func NewRabbitConnection(connectionString string,timeout int) (*RabbitConnection, error) {
	done := make(chan struct{}, 0)

	conn, err := amqp.DialConfig(connectionString,amqp.Config{
		Dial:amqp.DefaultDial(time.Second*time.Duration(timeout)),
	})

	if err != nil {
		return nil, err
	}
	newConnection := &RabbitConnection{
		done:       done,
		Connection: conn,
	}
	newConnection.waitClose()
	return newConnection, nil
}

func (f *RabbitConnection) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	return f.Connection.Close()
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
				break
			}
		}
	}()
}
