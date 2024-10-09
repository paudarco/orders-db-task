# Используем официальный образ Golang как базовый
FROM golang:1.17-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd

# Используем минимальный образ для запуска приложения
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем собранное приложение из предыдущего этапа
COPY --from=builder /app/main .
COPY --from=builder /app/config.json .

# Запускаем приложение
CMD ["./main"]
