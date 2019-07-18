package main

import "C"

import (
	"errors"
	"fmt"
	"github.com/arteev/rabbitmq-lib/src/logger"


	"log"
	"unsafe"
)

var logCommon *logger.Logger

var objects = map[uintptr]interface{}{}

type WrapError struct {
	e error
}


func putObject(obj interface{}) uintptr {
	key := uintptr(unsafe.Pointer(&obj))
	objects[key] = obj
	return key
}
func getObject(addr uintptr) interface{} {
	obj,ok := objects[addr]
	if !ok {
		return nil
	}
	return obj
}

func getObjectEx(addr uintptr) interface{} {
	return objects[addr]
}

type (
	Connection interface {}
	Channel interface {
		Testcall()
	}

)

//export Connect
func Connect()  {
	//return &test{}

	//C.CString(errors.New("123").Error())
	/*return struct {

	}{}*/
}

//export NewChannel
func NewChannel() uintptr {
	err := &WrapError{errors.New("test")}
	//rr := *(*iface)(unsafe.Pointer(&err))
	//adr:=uintptr(unsafe.Pointer(err))
	//fmt.Printf("%d %v\n",adr, reflect.ValueOf(err).Pointer())
	adr := putObject(err)
	return adr
}

//export FreeObject
func FreeObject(ptr uintptr) {
	delete(objects,ptr)
}

//export Disconnect
func Disconnect(Connection) {

}

//export CloseChannel
func CloseChannel(ptr uintptr) {
	w := getObject(ptr).(*WrapError)
	fmt.Printf("%d\n",ptr)
	fmt.Println(w.e)
}

//export InitLog
func InitLog(name string) error {
	if logCommon != nil {
		CloseLog()
	}
	newLog, err := logger.New(name)
	if err!= nil {
		return err
	}
	logCommon = newLog
	return nil
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
	if logCommon ==nil {
		return
	}
	logCommon.Close()
	logCommon = nil
}

//export IsError
func IsError(err interface{}) bool {
	if err == nil {
		return false
	}
	are, ok := err.(error)
	if !ok {
		return false
	}
	return are != nil
}

//export GetError
func GetError(err error) string {
	if err==nil {
		return ""
	}
	return err.Error()
}

func main() {

}
