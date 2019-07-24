package main

import (
	"errors"
	"github.com/arteev/rabbitmq-lib/src/logger"
	"github.com/arteev/rabbitmq-lib/src/rabbit"
	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"unsafe"
)

func TestMapArgs(t *testing.T) {
	m := MapArgs()
	assert.NotZero(t, m)
	obj, ok := getObjectEx(m)
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{}, obj)

	key := "param"
	value := "value"
	ok = MapArgsAdd(m, key, value)
	assert.True(t, ok)
	gotMap := getObject(m).(map[string]interface{})
	assert.Equal(t, map[string]interface{}{
		"param": "value",
	}, gotMap)

	ok = MapArgsAdd(0, "k", "v")
	assert.False(t, ok)

}

func TestObjects(t *testing.T) {

	obj := struct{}{}
	ptr := uintptr(unsafe.Pointer(&obj))
	putObject(ptr, obj)
	assert.Contains(t, objects, ptr)

	got := getObject(0)
	assert.Nil(t, got)

	got = getObject(ptr)
	assert.Equal(t, obj, got)

	_, ok := getObjectEx(0)
	assert.False(t, ok)

	got, ok = getObjectEx(ptr)
	assert.True(t, ok)
	assert.Equal(t, obj, got)

	FreeObject(ptr)
	assert.NotContains(t, objects, ptr)
}

func TestRabbit(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConnection := rabbit.NewMockConnection(ctrl)
	mockConnection.EXPECT().NotifyClose(gomock.Any()).AnyTimes()

	rabbitTest, err := rabbit.NewRabbitConnection(mockConnection, "", 10)
	assert.NoError(t, err)
	assert.NotNil(t, rabbitTest)
	ptrRabbit := uintptr(unsafe.Pointer(rabbitTest))
	putObject(ptrRabbit, rabbitTest)

	//Connected
	assert.False(t, Connected(0))
	assert.True(t, Connected(ptrRabbit))

	//Disconnect
	Disconnect(0)
	assert.True(t, Connected(ptrRabbit))
	mockConnection.EXPECT().Close()
	Disconnect(ptrRabbit)
	assert.False(t, Connected(ptrRabbit))

	//Channel
	rabbitTest, _ = rabbit.NewRabbitConnection(mockConnection, "", 10)
	ptrChannel := NewChannel(0)
	assert.Zero(t, ptrChannel)

	wantError := errors.New("channel error")
	mockConnection.EXPECT().Channel().Return(nil, wantError)
	ptrChannel = NewChannel(ptrRabbit)
	assert.Zero(t, ptrChannel)

	mockChannel := rabbit.NewMockChannel(ctrl)
	mockConnection.EXPECT().Channel().Return(mockChannel, nil)
	ptrChannel = NewChannel(ptrRabbit)
	assert.Equal(t, mockChannel, getObject(ptrChannel))

	CloseChannel(0)
	mockChannel.EXPECT().Close()
	CloseChannel(ptrChannel)

	ptrMap := MapArgs()
	MapArgsAdd(ptrMap, "k", "v")

	//ExchangeDeclare
	got := ExchangeDeclare(0, "", "", false, false, false, false, 0)
	assert.False(t, got)
	wantError = errors.New("exchange error")
	mockChannel.EXPECT().ExchangeDeclare("ex", "fanout", true, false, true, false,
		gomock.AssignableToTypeOf(map[string]interface{}{})).Return(wantError)
	got = ExchangeDeclare(ptrChannel, "ex", "fanout", true, false, true, false, ptrMap)
	assert.False(t, got)
	mockChannel.EXPECT().ExchangeDeclare("ex", "fanout", true, false, true, false,
		gomock.AssignableToTypeOf(map[string]interface{}{}))
	got = ExchangeDeclare(ptrChannel, "ex", "fanout", true, false, true, false, ptrMap)
	assert.True(t, got)

	//QueueDeclare
	got = QueueDeclare(0, "", false, false, false, false, 0)
	assert.False(t, got)
	wantError = errors.New("queue error")
	mockChannel.EXPECT().QueueDeclare("q1", true, false, true, false,
		gomock.AssignableToTypeOf(map[string]interface{}{})).Return(amqp.Queue{}, wantError)
	got = QueueDeclare(ptrChannel, "q1", true, false, true, false, ptrMap)
	assert.False(t, got)
	mockChannel.EXPECT().QueueDeclare("q1", true, false, true, false,
		gomock.AssignableToTypeOf(map[string]interface{}{})).Return(amqp.Queue{}, nil)
	got = QueueDeclare(ptrChannel, "q1", true, false, true, false, ptrMap)
	assert.True(t, got)

	//QueueBind
	got = QueueBind(0, "", "", "", false, 0)
	assert.False(t, got)
	wantError = errors.New("QueueBind error")
	mockChannel.EXPECT().QueueBind("n1", "n2", "n3", true,
		gomock.AssignableToTypeOf(map[string]interface{}{})).Return(wantError)
	got = QueueBind(ptrChannel, "n1", "n2", "n3", true, ptrMap)
	assert.False(t, got)
	mockChannel.EXPECT().QueueBind("n1", "n2", "n3", true,
		gomock.AssignableToTypeOf(map[string]interface{}{}))
	got = QueueBind(ptrChannel, "n1", "n2", "n3", true, ptrMap)
	assert.True(t, got)

	//Publish
	got = Publish(0, "", "", false, false, "", nil)
	assert.False(t, got)
	wantError = errors.New("Publish error")
	data := []byte("test")
	mockChannel.EXPECT().Publish("ex1", "key", true, true,
		amqp.Publishing{
			MessageId:    "id",
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         data,
		}).Return(wantError)
	got = Publish(ptrChannel, "ex1", "key", true, true, "id", data)
	assert.False(t, got)
	mockChannel.EXPECT().Publish("ex1", "key", true, true, gomock.Any())
	got = Publish(ptrChannel, "ex1", "key", true, true, "id", data)
	assert.True(t, got)
}

func TestLog(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	wantError := errors.New("log error")
	nameLog := "test.log"
	logger.OpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		assert.Equal(t, nameLog, name)
		return nil, wantError
	}
	got := InitLog(nameLog)
	assert.False(t, got)

	r,_,err:=os.Pipe()
	assert.NoError(t,err)

	logger.OpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return r, nil
	}
	got = InitLog(nameLog)
	assert.True(t, got)
	assert.NotNil(t, logCommon)
	assert.Equal(t, r, logCommon.GetOutput())

	PrintLog("test")

	CloseLog()
	assert.Nil(t,logCommon)
}
