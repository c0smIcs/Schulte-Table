BINARY_NAME=schulte
MAIN_PATH=./cmd/schulte/main.go

.PHONY: help
help:
	@echo "Доступные команды:"
	@echo "  make run   - Запустить приложение" 
	@echo "  make build - Собрать исполняемый файл" 
	@echo "  make clean - Удалить скомпилированный файл" 
	@echo "  make tidy  - Очистить и обновить зависимости (go mod tidy)" 

# Запуск приложения
.PHONY: run
run: build
	./.bin/$(BINARY_NAME) 2>&1 | jq -R -r 'fromjson? | .'

# Сборка приложения
.PHONY: build
build:
	@echo "Сборка..."
	go build -o ./.bin/$(BINARY_NAME) $(MAIN_PATH)

# очистка
.PHONY: clean
clean:
	@if [ -f $(BINARY_NAME) ] ; then rm $(BINARY_NAME); fi
	@echo "Очищено"

.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy