# build
FROM golang:1-bullseye as build
WORKDIR /build
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o cic-custodial -ldflags="-s -w" cmd/service/*.go

# main
FROM debian:bullseye-slim
ENV DEBIAN_FRONTEND=noninteractive
RUN set -x && \
    apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*
WORKDIR /service
COPY --from=build /build/cic-custodial .
COPY migrations migrations/
COPY config.toml .
COPY queries.sql .
CMD ["./cic-custodial"]
EXPOSE 5000
