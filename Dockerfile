FROM debian:11-slim

RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /cic-custodial

COPY cic-custodial .
COPY config.toml .

EXPOSE 5000
CMD ["./cic-dw"]
