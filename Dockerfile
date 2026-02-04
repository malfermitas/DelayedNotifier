# === Билд-стейдж для API ===
FROM golang:1.23-alpine AS api-builder

WORKDIR /app

# Установка delve для отладки
RUN go install github.com/go-delve/delve/cmd/dlv@v1.23.1

# Копируем модули
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Сборка API с поддержкой дебага
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o main ./cmd/app/main.go


# === Билд-стейдж для Worker ===
FROM golang:1.23-alpine AS worker-builder

WORKDIR /app

# Установка delve для отладки
RUN go install github.com/go-delve/delve/cmd/dlv@v1.23.1

# Копируем модули
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники
COPY . .

# Сборка Worker с поддержкой дебага
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="all=-N -l" -o main ./cmd/worker/main.go


# === Debug версия API ===
FROM alpine:latest AS debug-api

WORKDIR /root/

# Зависимости для delve
RUN apk add --no-cache libc6-compat

# Принудительный кэшбастинг
ARG CACHEBUST=1
RUN echo "Debug stage cache bust: $CACHEBUST"

# Копируем бинарь API, delve, конфиг и шаблоны
COPY --from=api-builder /app/main .
COPY --from=api-builder /go/bin/dlv /usr/local/bin/dlv
COPY config.yaml .
COPY .env .
COPY templates ./templates

# Очищаем кэш перед запуском
RUN sync

EXPOSE 8080 40000

# Команда запуска с delve
CMD ["dlv", "exec", "./main", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient"]


# === Debug версия Worker ===
FROM alpine:latest AS debug-worker

WORKDIR /root/

# Зависимости для delve
RUN apk add --no-cache libc6-compat

# Принудительный кэшбастинг
ARG CACHEBUST=1
RUN echo "Debug stage cache bust: $CACHEBUST"

# Копируем бинарь Worker, delve, конфиг и шаблоны
COPY --from=worker-builder /app/main .
COPY --from=worker-builder /go/bin/dlv /usr/local/bin/dlv
COPY config_worker.yaml ./config_worker.yaml
COPY .env .
COPY templates ./templates

# Очищаем кэш перед запуском
RUN sync

EXPOSE 40001 40002

# Команда запуска с delve
CMD ["dlv", "exec", "./main", "--listen=:40002", "--headless=true", "--api-version=2", "--accept-multiclient"]


# === Релизная версия API ===
FROM alpine:latest AS release-api

WORKDIR /root/

# Копируем только бинарь и нужные данные
COPY --from=api-builder /app/main .
COPY config.yaml .
COPY templates ./templates

EXPOSE 8080

CMD ["./main"]


# === Релизная версия Worker ===
FROM alpine:latest AS release-worker

WORKDIR /root/

# Копируем только бинарь и нужные данные
COPY --from=worker-builder /app/main .
COPY config_worker.yaml ./config.yaml
COPY templates ./templates

EXPOSE 40001

CMD ["./main"]
