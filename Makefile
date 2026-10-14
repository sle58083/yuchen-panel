.PHONY: backend agent test docker

backend:
	cd backend && go run ./cmd/server

agent:
	cd agent && go run -ldflags "-X main.version=$(AGENT_VERSION)" ./cmd/agent

test:
	cd backend && go test ./...
	cd agent && go test ./...

docker:
	docker compose up -d --build

AGENT_VERSION := $(shell cut -d- -f1 VERSION)
.PHONY: build-agent
build-agent:
	cd agent && CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=$(AGENT_VERSION)" -o ../bin/yuchen-agent-linux-amd64 ./cmd/agent
