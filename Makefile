# Переменные проекта
BINARY_NAME=schulte
MAIN_PATH=./cmd/schulte/main.go
BIN_DIR=./.bin

# Цвета для вывода (для красоты)
YELLOW := \033[33m
RESET  := \033[0m

.PHONY: help
help: ## Показать справку
	@echo "$(YELLOW)Доступные команды:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: run
run: build ## Собрать и запустить приложение локально с форматированием логов
	$(BIN_DIR)/$(BINARY_NAME) 2>&1 | jq -R -r 'fromjson? | .'

.PHONY: build
build: ## Собрать исполняемый файл
	@echo "$(YELLOW)Сборка...$(RESET)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)

.PHONY: clean
clean: ## Удалить скомпилированные файлы
	@rm -rf $(BIN_DIR)
	@echo "$(YELLOW)Очищено$(RESET)"

.PHONY: tidy
tidy: ## Форматировать код и обновить зависимости
	go fmt ./...
	go mod tidy

# --- Docker команды ---

.PHONY: docker-up
docker-up: ## Запустить проект в Docker (сборка + подъем)
	docker compose up --build

.PHONY: docker-down
docker-down: ## Остановить контейнеры
	docker compose down

.PHONY: docker-reset
docker-reset: ## Остановить контейнеры и удалить тома (очистка БД)
	docker compose down -v

# Алиасы
docker: docker-up

