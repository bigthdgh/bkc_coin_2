# 🚀 BKC Coin 10K Optimized Dockerfile
FROM golang:1.21-alpine AS builder

# 📦 Установка зависимостей
RUN apk add --no-cache git ca-certificates tzdata

# ⚡ Оптимизации сборки
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 📁 Рабочая директория
WORKDIR /app

# 📋 Копирование модулей
COPY go.mod go.sum ./
RUN go mod download -x

# 📦 Копирование исходного кода
COPY . .

# 🔨 Сборка оптимизированного бинарника
RUN go build -ldflags="-w -s" -o bkc-10k ./cmd/server/main_10k_optimized.go

# 🚀 Финальный образ
FROM alpine:latest

# 📦 Установка runtime зависимостей
RUN apk --no-cache add ca-certificates tzdata

# 📁 Создание пользователя
RUN addgroup -g 1001 -S bkc && \
    adduser -u 1001 -S bkc -G bkc

# 📁 Рабочая директория
WORKDIR /app

# 📋 Копирование бинарника
COPY --from=builder /app/bkc-10k .

# 📁 Копирование статических файлов
COPY --from=builder /app/webapp ./webapp

# 📁 Создание необходимых директорий
RUN mkdir -p /app/logs /app/cache && \
    chown -R bkc:bkc /app

# 🔄 Переключение на пользователя
USER bkc

# 🚀 Запуск
EXPOSE 10000

# 🏥 Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:10000/health || exit 1

# 🚀 Команда запуска
CMD ["./bkc-10k"]
