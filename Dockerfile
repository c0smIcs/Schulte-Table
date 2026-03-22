# =============================================================================
# Schulte Table — образ приложения (multi-stage)
# =============================================================================
#
# Сборка:
#   docker build -t schulte-app .
#
# Запуск только приложения (без БД — для отладки; обычно используй docker compose):
#   docker run --rm -p 8080:8080 schulte-app
#
# Рабочая директория в рантайме: /root
# Конфиг:        ./configs/config.yml  (в compose монтируется config.docker.yml)
# Статика/UI:    ./ui
# Миграции SQL:  ./migrations (копируются в образ; применение к БД — через compose/initdb)
#
# Go: версия образа должна быть >= строки `go` в go.mod.
# Сборка с CGO_ENABLED=0 даёт статический бинарник, удобный для Alpine.
#
# Если сборка падает с EOF / timeout при загрузке манифеста с Docker Hub:
#   - повтори `docker compose build --no-cache` позже (часто сбой сети);
#   - `docker pull alpine:latest` и `docker pull golang:1.26-alpine` вручную;
#   - проверь VPN/прокси/DNS; при лимите Hub: `docker login`.
# =============================================================================

# --- этап сборки ---
FROM golang:1.26-alpine AS builder
WORKDIR /schulte

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o schulte-app ./cmd/schulte/main.go

# --- финальный образ ---
FROM alpine:latest
WORKDIR /root/

COPY --from=builder /schulte/schulte-app .
COPY --from=builder /schulte/configs ./configs
COPY --from=builder /schulte/ui ./ui
COPY --from=builder /schulte/migrations ./migrations

EXPOSE 8080
CMD ["./schulte-app"]
