package main

import "C"

import (
	"github.com/arteev/rabbitmq-lib/src/logger"
	"github.com/arteev/rabbitmq-lib/src/rabbit"
	"log"
	"sync"
	"unsafe"

	"github.com/streadway/amqp"
)

var (
	logCommon *logger.Logger
	objects   = map[uintptr]interface{}{}
	muObjects sync.RWMutex
)

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
func Connect(connectionString string, timeout int) uintptr {
	conn, err := rabbit.NewRabbitConnection(nil,connectionString, timeout)
	if err != nil {
		log.Println("Connect error:", err)
		return 0
	}
	log.Println("Connected ", connectionString)
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
	conn := obj.(*rabbit.RabbitConnection)
	return conn.Connected()
}

//export NewChannel
func NewChannel(connPtr uintptr) uintptr {
	obj, ok := getObjectEx(connPtr)
	if !ok {
		return 0
	}
	conn := obj.(*rabbit.RabbitConnection)
	log.Println("NewChannel connection:", connPtr)
	channel, err := conn.Channel()
	if err != nil {
		log.Println("NewChannel", connPtr, "Error:", err)
		return 0
	}
	ptr := uintptr(unsafe.Pointer(&channel))
	putObject(ptr, channel)
	return ptr
}

//export ExchangeDeclare
func ExchangeDeclare(channelPtr uintptr,
	name, kind string, durable, autoDelete, internal, noWait bool, args uintptr) bool {
	obj, ok := getObjectEx(channelPtr)
	if !ok {
		return false
	}
	channel := obj.(rabbit.Channel)
	argsRaw, ok := getObjectEx(args)
	var mArgs map[string]interface{}
	if ok {
		mArgs = argsRaw.(map[string]interface{})
	}
	log.Println("ExchangeDeclare", channelPtr, name, kind, mArgs)
	err := channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, mArgs)
	if err != nil {
		log.Println("ExchangeDeclare ", channelPtr, "Error:", err)
		return false
	}
	return true
}

//export QueueDeclare
func QueueDeclare(channelPtr uintptr, name string, durable, autoDelete, exclusive, noWait bool, args uintptr) bool {
	obj, ok := getObjectEx(channelPtr)
	if !ok {
		return false
	}

	argsRaw, ok := getObjectEx(args)
	var mArgs map[string]interface{}
	if ok {
		mArgs = argsRaw.(map[string]interface{})
	}
	channel := obj.(rabbit.Channel)
	log.Println("QueueDeclare", channelPtr, name, mArgs)
	_, err := channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, mArgs)
	if err != nil {
		log.Println("QueueDeclare", channelPtr, "Error:", err)
		return false
	}
	return true
}

//export QueueBind
func QueueBind(channelPtr uintptr, name, key, exchange string, noWait bool, args uintptr) bool {
	obj, ok := getObjectEx(channelPtr)
	if !ok {
		return false
	}
	channel := obj.(rabbit.Channel)
	argsRaw, ok := getObjectEx(args)
	var mArgs map[string]interface{}
	if ok {
		mArgs = argsRaw.(map[string]interface{})
	}
	log.Println("QueueBind", channelPtr, name, key, exchange)
	err := channel.QueueBind(name, key, exchange, noWait, mArgs)
	if err != nil {
		log.Println("QueueBind", channelPtr, "Error:", err)
		return false
	}

	return true
}

//export Publish
func Publish(channelPtr uintptr, exchange, key string, mandatory, immediate bool, messageID string, msg []byte) bool {
	obj, ok := getObjectEx(channelPtr)
	if !ok {
		return false
	}
	channel := obj.(rabbit.Channel)
	err := channel.Publish(exchange, key, mandatory, immediate, amqp.Publishing{
		MessageId:    messageID,
		DeliveryMode: amqp.Persistent,
		ContentType:  "text/plain",
		Body:         msg,
	})
	if err != nil {
		log.Println("Publish", channelPtr, "Error:", err)
		return false
	}
	log.Printf("Publish channel:%v, ecxhange:%q, key:%q, id:%q, message:%s\n", channelPtr, exchange, key,
		messageID, string(msg))
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
	conn := obj.(*rabbit.RabbitConnection)
	conn.Close()
	log.Println("Disconnected", ptr)
}

//export CloseChannel
func CloseChannel(ptr uintptr) {
	obj, ok := getObjectEx(ptr)
	if !ok {
		return
	}
	channel := obj.(rabbit.Channel)
	channel.Close()
	log.Println("Channel closed", ptr)
}

//export InitLog
func InitLog(name string) bool {
	if logCommon != nil {
		CloseLog()
	}
	newLog, err := logger.New(name)
	if err != nil {
		log.Println(name, err)
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

//export MapArgs
func MapArgs() uintptr {
	m := map[string]interface{}{}
	ptr := uintptr(unsafe.Pointer(&m))
	putObject(ptr, m)
	return ptr
}

//export MapArgsAdd
func MapArgsAdd(ptr uintptr, key string, value string) bool {
	obj, ok := getObjectEx(ptr)
	if !ok {
		return false
	}
	m := obj.(map[string]interface{})
	//NOTE: copy strings as the caller frees memory immediately
	cKey := make([]byte, len(key))
	cValue := make([]byte, len(value))
	copy(cKey, key)
	copy(cValue, value)
	m[string(cKey)] = string(cValue)
	return true
}

func main() {

}
