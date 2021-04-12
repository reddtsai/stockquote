.PHONY:

app-mining:
	go run ./cmd/mining/main.go

app-exchange:
	go run ./cmd/exchange/main.go

unit-test:
	go test ./pkg/twse -v

up:
	docker build -t app-exchange -f build/app-exchange/Dockerfile . --no-cache
	docker build -t app-mining -f build/app-mining/Dockerfile . --no-cache
	docker-compose -f build/docker-compose.yaml -p dev up -d

down:
	docker-compose -f build/docker-compose.yaml -p dev down