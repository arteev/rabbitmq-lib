package main

import "C"

import (
	"fmt"
	"github.com/arteev/rabbitmq-lib/src/logger"
	"log"
	"sync"
	"unsafe"

	"github.com/streadway/amqp"
)

var logCommon *logger.Logger
var objects = map[uintptr]interface{}{}
var muObjects sync.RWMutex

func putObject(key uintptr, obj interface{}) {
	muObjects.Lock()
	defer muObjects.Unlock()
	objects[key] = obj
}
func getObject(addr uintptr) interface{} {
	muObjects.RLock()
	defer muObjects.RUnlock()
	obj, ok := objects[addr]
	if !ok {
		return nil
	}
	return obj
}

func getObjectEx(addr uintptr) (interface{}, bool) {
	muObjects.RLock()
	defer muObjects.RUnlock()
	obj, ok := objects[addr]
	return obj, ok
}

//export Connect
func Connect(connectionString string,timeout int) uintptr {
	conn, err := NewRabbitConnection(connectionString,timeout)
	if err != nil {
		log.Println(err)
		return 0
	}
	ptr := uintptr(unsafe.Pointer(conn))
	putObject(ptr, conn)
	return ptr
}

//export Connected
func Connected(connPtr uintptr) bool {
	obj, ok := getObjectEx(connPtr)
	if !ok {
		return false
	}
	conn := obj.(*RabbitConnection)
	return conn.Connected()
}


//export NewChannel
func NewChannel(connPtr uintptr) uintptr {
	obj, ok := getObjectEx(connPtr)
	if !ok {
		return 0
	}
	conn := obj.(*RabbitConnection)
	channel, err := conn.Channel()
	if err != nil {
		log.Println(err)
		return 0
	}
	ptr := uintptr(unsafe.Pointer(channel))
	putObject(ptr, channel)
	//log.Println(ptr, reflect.ValueOf(channel).Pointer())
	return ptr
}

//export ExchangeDeclare
func ExchangeDeclare(channelPtr uintptr,
	name, kind string, durable, autoDelete, internal, noWait bool) bool {
	obj, ok := getObjectEx(channelPtr)
	if !ok {
		return false
	}
	channel := obj.(*amqp.Channel)
	err := channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, nil)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println("ExchangeDeclare",name,kind)
	return true
}

//export QueueDeclare
func QueueDeclare(channelPtr uintptr, name string, durable, autoDelete, exclusive, noWait bool) bool {
	obj, ok := getObjectEx(channelPtr)
	if !ok {
		return false
	}
	channel := obj.(*amqp.Channel)
	_, err := channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, nil)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println("QueueDeclare",name)
	return true
}

//export QueueBind
func QueueBind(channelPtr uintptr, name, key, exchange string, noWait bool) bool {
	obj, ok := getObjectEx(channelPtr)
	if !ok {
		return false
	}
	channel := obj.(*amqp.Channel)
	err := channel.QueueBind(name, key, exchange, noWait, nil)
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println("QueueBind",name,key,exchange)
	return true
}

//export Publish
func Publish(channelPtr uintptr, exchange, key string, mandatory, immediate bool, msg []byte) bool {
	obj, ok := getObjectEx(channelPtr)
	if !ok {
		return false
	}
	channel := obj.(*amqp.Channel)
	err := channel.Publish(exchange, key, mandatory, immediate, amqp.Publishing{
		ContentType: "text/plain",
		Body:        msg,
	})
	if err != nil {
		log.Println(err)
		return false
	}
	log.Println("QueueBind",exchange,key,string(msg))
	return true
}

//export FreeObject
func FreeObject(ptr uintptr) {
	delete(objects, ptr)
}

//export Disconnect
func Disconnect(ptr uintptr) {
	obj, ok := getObjectEx(ptr)
	if !ok {
		return
	}
	conn := obj.(*RabbitConnection)
	conn.Close()
}

//export CloseChannel
func CloseChannel(ptr uintptr) {
	obj, ok := getObjectEx(ptr)
	if !ok {
		return
	}
	channel := obj.(*amqp.Channel)
	channel.Close()
}

//export InitLog
func InitLog(name string) bool {
	if logCommon != nil {
		CloseLog()
	}
	newLog, err := logger.New(name)
	if err != nil {
		fmt.Println(name, err)
		return false
	}
	newLog.ApplyToStdLog()
	logCommon = newLog
	return true
}

//export PrintLog
func PrintLog(s string) {
	if logCommon == nil {
		return
	}
	log.Println(s)
}

//export CloseLog
func CloseLog() {
	if logCommon == nil {
		return
	}
	logCommon.Close()
	logCommon = nil
}

func main() {

}
