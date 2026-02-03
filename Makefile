# Makefile para Mikrom Go API

# Variables
APP_NAME=mikrom-api
VERSION?=1.0.0
BUILD_DIR=bin
COVERAGE_DIR=coverage
MAIN_PATH=cmd/api/main.go

# Configuración
.DEFAULT_GOAL := help
.PHONY: help run dev build test test-coverage test-verbose clean docker-up docker-down docker-restart install lint fmt vet check deps migrate-up migrate-down

## help: Muestra esta ayuda
help:
	@echo 'Uso:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run: Ejecuta la aplicación
run:
	@echo "Iniciando aplicación..."
	go run $(MAIN_PATH)

## dev: Ejecuta la aplicación con hot-reload (requiere air: go install github.com/cosmtrek/air@latest)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air no está instalado. Instálalo con: go install github.com/cosmtrek/air@latest"; \
		echo "O ejecuta: make run"; \
	fi

## build: Compila la aplicación para producción
build:
	@echo "Compilando $(APP_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w -X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)
	@echo "Binario creado en $(BUILD_DIR)/$(APP_NAME)"

## build-linux: Compila para Linux (útil si estás en Mac/Windows)
build-linux:
	@echo "Compilando para Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME)-linux $(MAIN_PATH)

## build-all: Compila para múltiples plataformas
build-all:
	@echo "Compilando para múltiples plataformas..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "Binarios creados en $(BUILD_DIR)/"

## test: Ejecuta todos los tests
test:
	@echo "Ejecutando tests..."
	go test ./... -race

## test-verbose: Ejecuta los tests con output detallado
test-verbose:
	@echo "Ejecutando tests (verbose)..."
	go test ./... -v -race

## test-short: Ejecuta tests rápidos (sin tests de integración)
test-short:
	@echo "Ejecutando tests cortos..."
	go test ./... -short

## test-coverage: Ejecuta los tests y genera reporte de cobertura
test-coverage:
	@echo "Ejecutando tests con cobertura..."
	@mkdir -p $(COVERAGE_DIR)
	go test ./... -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic
	go tool cover -func=$(COVERAGE_DIR)/coverage.out
	@echo "\n✓ Para ver el reporte HTML ejecuta: make coverage-html"

## coverage-html: Genera reporte HTML de cobertura
coverage-html:
	@if [ ! -f $(COVERAGE_DIR)/coverage.out ]; then \
		echo "Ejecuta primero: make test-coverage"; \
		exit 1; \
	fi
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "✓ Reporte generado en $(COVERAGE_DIR)/coverage.html"

## bench: Ejecuta benchmarks
bench:
	@echo "Ejecutando benchmarks..."
	go test ./... -bench=. -benchmem

## vet: Ejecuta go vet
vet:
	@echo "Ejecutando go vet..."
	go vet ./...

## fmt: Formatea el código
fmt:
	@echo "Formateando código..."
	go fmt ./...

## fmt-check: Verifica que el código esté formateado
fmt-check:
	@echo "Verificando formato del código..."
	@test -z "$$(gofmt -l .)" || (echo "Los siguientes archivos necesitan formato:" && gofmt -l . && exit 1)

## lint: Ejecuta el linter (requiere golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Ejecutando linter..."; \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint no está instalado."; \
		echo "Instálalo desde: https://golangci-lint.run/usage/install/"; \
	fi

## check: Ejecuta todas las verificaciones (fmt, vet, lint, test)
check: fmt-check vet lint test
	@echo "✓ Todas las verificaciones pasaron"

## install: Instala dependencias
install:
	@echo "Instalando dependencias..."
	go mod download
	go mod tidy
	@echo "✓ Dependencias instaladas"

## deps: Actualiza dependencias
deps:
	@echo "Actualizando dependencias..."
	go get -u ./...
	go mod tidy
	@echo "✓ Dependencias actualizadas"

## tidy: Limpia y organiza go.mod
tidy:
	@echo "Limpiando go.mod..."
	go mod tidy
	@echo "✓ go.mod limpio"

## clean: Limpia archivos generados
clean:
	@echo "Limpiando archivos generados..."
	rm -rf $(BUILD_DIR)
	rm -rf $(COVERAGE_DIR)
	rm -f coverage.out coverage.html
	go clean -cache -testcache -modcache
	@echo "✓ Limpieza completada"

## docker-up: Inicia PostgreSQL con Docker Compose
docker-up:
	@echo "Iniciando PostgreSQL..."
	docker compose up -d
	@echo "✓ PostgreSQL iniciado en puerto 5432"

## docker-down: Detiene y elimina contenedores Docker
docker-down:
	@echo "Deteniendo Docker Compose..."
	docker-compose down
	@echo "✓ Contenedores detenidos"

## docker-down-v: Detiene contenedores y elimina volúmenes
docker-down-v:
	@echo "Deteniendo Docker Compose y eliminando volúmenes..."
	docker-compose down -v
	@echo "✓ Contenedores y volúmenes eliminados"

## docker-restart: Reinicia los servicios Docker
docker-restart: docker-down docker-up

## docker-logs: Muestra logs de PostgreSQL
docker-logs:
	docker-compose logs -f postgres

## docker-ps: Muestra estado de contenedores
docker-ps:
	docker-compose ps

## migrate-up: Ejecuta migraciones de base de datos (si existen)
migrate-up:
	@echo "Las migraciones se ejecutan automáticamente al iniciar la aplicación"

## migrate-down: Revierte migraciones (implementar según necesidad)
migrate-down:
	@echo "No implementado aún"

## db-shell: Abre shell de PostgreSQL
db-shell:
	docker-compose exec postgres psql -U postgres -d mikrom

## seed: Carga datos de prueba en la base de datos
seed:
	@echo "Cargando datos de prueba..."
	@echo "No implementado aún"

## watch: Observa cambios y ejecuta tests automáticamente
watch:
	@if command -v watchexec > /dev/null; then \
		watchexec -e go -r "make test"; \
	else \
		echo "watchexec no está instalado."; \
		echo "Instálalo desde: https://github.com/watchexec/watchexec"; \
	fi

## setup: Configuración inicial del proyecto
setup: install docker-up
	@echo "Esperando a que PostgreSQL esté listo..."
	@sleep 3
	@echo "✓ Configuración completada"
	@echo ""
	@echo "Siguiente paso: cp .env.example .env y configura tus variables"
	@echo "Luego ejecuta: make run"

## all: Ejecuta verificaciones, tests y compila
all: check test build
	@echo "✓ Build completo exitoso"

## info: Muestra información del proyecto
info:
	@echo "Proyecto: $(APP_NAME)"
	@echo "Versión: $(VERSION)"
	@echo "Go version: $$(go version)"
	@echo "Directorio build: $(BUILD_DIR)"
	@echo "Archivos Go: $$(find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "Paquetes: $$(go list ./... | wc -l)"
