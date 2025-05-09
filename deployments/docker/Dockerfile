FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o wallet-service ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/wallet-service .
COPY --from=builder /app/config.env .

EXPOSE 8080

# Команда для запуска миграций и приложения
CMD ["./wallet-service"]
