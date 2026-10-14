FROM golang:1.23 AS builder
WORKDIR /src
RUN apt-get update \
    && apt-get install -y --no-install-recommends bash curl jq python3 unzip \
    && rm -rf /var/lib/apt/lists/*
COPY backend/ ./backend/
COPY install.sh ./install.sh
COPY VERSION ./VERSION
WORKDIR /src/backend
RUN env PATH=/usr/local/go/bin:/bin /usr/local/go/bin/go test -run '^TestSubscriptionFormatRouting$' ./internal/api
RUN go test -v ./... && go build -o /out/yuchen-panel-api ./cmd/server

FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && update-ca-certificates \
    && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /app/data
COPY --from=builder /out/yuchen-panel-api /app/yuchen-panel-api
EXPOSE 8088
CMD ["/app/yuchen-panel-api"]
