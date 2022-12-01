FROM debian:11-slim

ENV DEBIAN_FRONTEND=noninteractive
RUN set -x && \
    apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/cache/apt/archives /var/lib/apt/lists/*

WORKDIR /service

COPY cic-custodial .
COPY config.toml .

CMD ["./cic-custodial"]

EXPOSE 5000
