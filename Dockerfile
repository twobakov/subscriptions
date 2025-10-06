# этап сборки
FROM golang:1.24 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# сборка бинарика
RUN GOOS=linux GOARCH=amd64 go build -o todo-app ./cmd/api/main.go

# образ
FROM debian:bookworm-slim
WORKDIR /app

# копи бинаринк
COPY --from=builder /app/todo-app /app/todo-app

COPY --from=builder /app/config /app/config

COPY --from=builder /app/.env /app/.env

# запуск
CMD ["/app/todo-app"]
