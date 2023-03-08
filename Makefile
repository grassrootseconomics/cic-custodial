BIN := cic-custodial
BUILD_CONF := CGO_ENABLED=1 GOOS=linux GOARCH=amd64
BUILD_COMMIT := $(shell git rev-parse --short HEAD 2> /dev/null)

.PHONY: build

clean:
	rm ${BIN}

build:
	${BUILD_CONF} go build -ldflags="-X main.build=${BUILD_COMMIT} -s -w" -o ${BIN} cmd/service/*

run:
	${BUILD_CONF} go run cmd/service/*

run-debug:
	${BUILD_CONF} go run cmd/service/* -debug
