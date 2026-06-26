.PHONY: build run-server test clean deploy restart logs status db-up db-down

APP_NAME=memory-brain
SERVER_BIN=bin/server
CLI_BIN=bin/memory

build:
	mkdir -p bin
	go build -o $(SERVER_BIN) cmd/server/main.go
	go build -o $(CLI_BIN) cmd/cli/main.go

run-server:
	./$(SERVER_BIN)

dev-server:
	go run cmd/server/main.go

test:
	go test -v ./...

clean:
	rm -rf bin/

restart:
	pm2 restart $(APP_NAME)

logs:
	pm2 logs $(APP_NAME)

status:
	pm2 status $(APP_NAME)

deploy:
	git pull
	$(MAKE) build
	pm2 restart $(APP_NAME)
	pm2 save

# Solo para desarrollo local, no usar en server productivo
db-up:
	docker-compose up -d

db-down:
	docker-compose down
