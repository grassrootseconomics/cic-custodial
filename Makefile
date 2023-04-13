BIN := cic-custodial
BUILD_CONF := CGO_ENABLED=1 GOOS=linux GOARCH=amd64
BUILD_COMMIT := $(shell git rev-parse --short HEAD 2> /dev/null)

.PHONY: build run run-debug docs

clean:
	rm ${BIN}

build:
	${BUILD_CONF} go build -ldflags="-X main.build=${BUILD_COMMIT} -s -w" -o ${BIN} cmd/service/*

docs:
	swag init --dir internal/api/ -g swagger.go

run:
	${BUILD_CONF} go run cmd/service/*

run-debug:
	${BUILD_CONF} go run cmd/service/* -debug
