package logger

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	f *os.File
}

func New(fileName string) (*Logger, error) {
	f,err := os.OpenFile(fileName,os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil,err
	}
	newLogger := &Logger{
		f:f,
	}
	return newLogger, nil
}

func (l *Logger) GetOutput() io.Writer {
	return l.f
}

func (l *Logger) ApplyToStdLog() {
	log.SetOutput(l.GetOutput())
}

func (l *Logger) Close() error {
	return l.f.Close()
}
