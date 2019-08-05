package rabbit

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

func TestNewRabbitConnection(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConnection := NewMockConnection(ctrl)
	mockConnection.EXPECT().NotifyClose(gomock.Any()).AnyTimes()

	_, err := NewRabbitConnection(nil, "amqps://0.0.0.0:1", 10)
	assert.EqualError(t, err, "dial tcp 0.0.0.0:1: connect: connection refused")

	rabbit, err := NewRabbitConnection(mockConnection, "", 10)
	assert.NoError(t, err)
	assert.NotNil(t, rabbit)

	//channel
	wantError := errors.New("channel error")
	mockConnection.EXPECT().Channel().Return(nil, wantError)
	_, err = rabbit.Channel()
	assert.EqualError(t, err, wantError.Error())

	mockChannel := NewMockChannel(ctrl)
	mockConnection.EXPECT().Channel().Return(mockChannel, nil)
	got, err := rabbit.Channel()
	assert.NoError(t, err)
	assert.Equal(t, mockChannel, got)

	//connected, Close
	assert.True(t, rabbit.Connected())
	wantError = errors.New("close error")
	mockConnection.EXPECT().Close().Return(wantError)
	err = rabbit.Close()
	assert.EqualError(t, err, wantError.Error())
	assert.False(t, rabbit.Connected())

}

func TestRabbitConnectionWaitClose(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockConnection := NewMockConnection(ctrl)
	var chError chan *amqp.Error
	mockConnection.EXPECT().NotifyClose(gomock.Any()).Do(func(receiver chan *amqp.Error) chan *amqp.Error {
		chError = receiver
		return receiver
	}).AnyTimes()

	rabbit, err := NewRabbitConnection(mockConnection, "", 10)
	assert.NoError(t, err)
	assert.NotNil(t, rabbit)

	time.Sleep(time.Millisecond * 300)
	assert.True(t, rabbit.Connected())
	chError <- amqp.ErrClosed
	time.Sleep(time.Millisecond * 200)
	assert.False(t, rabbit.Connected())

	//check done
	rabbit, err = NewRabbitConnection(mockConnection, "", 10)
	assert.NoError(t, err)
	assert.NotNil(t, rabbit)
	time.Sleep(time.Millisecond * 300)
	assert.True(t, rabbit.Connected())
	mockConnection.EXPECT().Close()
	rabbit.Close()
	time.Sleep(time.Millisecond * 100)
	assert.False(t, rabbit.Connected())
	_, ok := <-rabbit.done
	assert.False(t, ok)

}
