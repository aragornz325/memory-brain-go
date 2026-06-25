.PHONY: build run-server run-db test migrate-status migrate-up migrate-down clean

# Compilar binarios
build:
	mkdir -p bin
	go build -o bin/server cmd/server/main.go
	go build -o bin/memory cmd/cli/main.go

# Ejecutar el servidor HTTP localmente
run-server:
	go run cmd/server/main.go

# Levantar base de datos PostgreSQL local
run-db:
	docker-compose up -d

# Detener base de datos
stop-db:
	docker-compose down

# Ejecutar pruebas unitarias
test:
	go test -v ./...

# Limpiar binarios
clean:
	rm -rf bin/
