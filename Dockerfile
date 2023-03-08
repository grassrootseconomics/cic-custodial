FROM golang:1-bullseye as build

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

ENV BUILD_COMMIT=${BUILD_COMMIT}

WORKDIR /build

COPY go.* .
RUN go mod download

COPY . .
RUN go build -o cic-custodial -ldflags="-X main.build=${BUILD_COMMIT} -s -w" cmd/service/*

FROM debian:bullseye-slim

ENV DEBIAN_FRONTEND=noninteractive

WORKDIR /service

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/cic-custodial .
COPY migrations migrations/
COPY config.toml .
COPY queries.sql .

EXPOSE 5000

CMD ["./cic-custodial"]
