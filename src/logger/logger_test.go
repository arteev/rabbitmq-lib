package logger

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

type mockFile struct {
	InvokedClose bool
}

func (f *mockFile) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func (f *mockFile) Close() error {
	f.InvokedClose = true
	return errors.New("close error")
}

func TestLogger(t *testing.T) {

	invoked := false
	wantError := errors.New("open error")
	testFile := "test.log"

	OpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		invoked = true
		assert.Equal(t, testFile, name)
		assert.Equal(t, os.FileMode(0666), perm)
		assert.Equal(t, 1090, flag)
		return nil, wantError
	}

	//error
	_, err := New(testFile)
	assert.EqualError(t, err, wantError.Error())
	assert.True(t, invoked)

	//normal
	want := os.Stdout
	OpenFile = func(name string, flag int, perm os.FileMode) (*os.File, error) {
		return want, nil
	}
	got, err := New(testFile)
	assert.NotNil(t, got)
	assert.Equal(t, want, got.f)
	assert.Equal(t, got.GetOutput(), got.f)

	mock := &mockFile{}
	got.f = mock
	err = got.Close()
	assert.True(t, mock.InvokedClose)
	assert.EqualError(t, err, "close error")

	got.ApplyToStdLog()
	err = log.Output(1, "test")
	assert.EqualError(t, err, "write error")

}
