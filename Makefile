VERSION=$(shell git describe --tags --abbrev=0 || echo "1.0.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
GITHEAD=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-w -s -X main.Version=${VERSION} -X main.DateBuild=${BUILD_TIME}  -X main.GitHash=${GITHEAD}"
OUTFILE=bin/librmq
SRC=./src
FILES=${SRC}/main.go
PKG_NAME=github.com/arteev/rabbitmq-lib

ifeq ($(OS), Windows_NT)
	OUTFILE=bin/librmq.dll
else
	OUTFILE=bin/librmq.so
endif


default: test build

#lint

build:
	@echo " >  Building ..."
	@GO111MODULE=on CGO_ENABLED=1 GODEBUG=cgocheck=2 go build ${LDFLAGS} -o ${OUTFILE} -buildmode=c-shared ${FILES}

run:
	@echo " >  Running ..."
	@GO111MODULE=on go run ${LDFLAGS} ${FILES}

test:
	@echo " >  Running tests ..."
	@GO111MODULE=on go test -v ${SRC}/...