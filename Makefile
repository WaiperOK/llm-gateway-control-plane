.PHONY: run test integration lint fmt docker-build docker-up docker-down

run:
	go run ./cmd/gateway

test:
	go test ./...

integration:
	go test ./tests/integration -v

lint:
	go vet ./...

fmt:
	gofmt -w $$(rg --files -g '*.go')

docker-build:
	docker build -t llm-gateway-control-plane:local .

docker-up:
	docker compose up --build -d

docker-down:
	docker compose down
