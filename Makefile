VERSION=$(shell git describe --tags --abbrev=0 || echo "1.0.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
GITHEAD=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-w -s -X main.Version=${VERSION} -X main.DateBuild=${BUILD_TIME}  -X main.GitHash=${GITHEAD}"
OUTFILE=bin/librmq
SRC=./src
FILES=${SRC}/main.go
PKG_NAME=github.com/arteev/rabbitmq-lib

ifeq ($(OS), Windows_NT)
	OUTFILE=bin/rmq.dll
else
	OUTFILE=bin/librmq.so
endif


default: test build

#lint

build:
	@echo " >  Building ..."
	@GO111MODULE=on CGO_ENABLED=1 GODEBUG=cgocheck=2 go build ${LDFLAGS} -o ${OUTFILE} -buildmode=c-shared ${FILES}

test:
	@echo " >  Running tests ..."
	@GO111MODULE=on go test -v ${SRC}/...


cover:
	@echo " >  Running tests with coverage..."
	@GO111MODULE=on go test -coverprofile=cover.out `go list ./... | grep -v mock` && grep -v mock cover.out > coverclean.out  &&  go tool cover -func=coverclean.out
	@rm -f cover.out
	@rm -f coverclean.out

mock-rabbit:
	mockgen  -package=rabbit -self_package=${PKG_NAME}/src/rabbit ${PKG_NAME}/src/rabbit Channel > src/rabbit/_mockchannel.go
	rm -f src/rabbit/mockchannel.go
	mv src/rabbit/_mockchannel.go src/rabbit/mockchannel.go

	mockgen  -package=rabbit -self_package=${PKG_NAME}/src/rabbit  ${PKG_NAME}/src/rabbit Connection > src/rabbit/_mockrmq.go
	rm -f src/rabbit/mockrmq.go
	mv src/rabbit/_mockrmq.go src/rabbit/mockrmq.go