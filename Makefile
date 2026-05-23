APP_NAME=opencops
MAIN_PATH=./cmd/opencops

.PHONY: run
run:
	go run $(MAIN_PATH)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: build
build:
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

.PHONY: docker-up
docker-up:
	docker compose up -d

.PHONY: docker-down
docker-down:
	docker compose down

.PHONY: docker-logs
docker-logs:
	docker compose logs -f

.PHONY: test
test:
	go test ./...