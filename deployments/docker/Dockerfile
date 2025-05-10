FROM golang:1.20-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

ENV GOPROXY=direct
RUN go mod tidy
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o wallet-service ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/wallet-service .
COPY --from=builder /app/config.env .

EXPOSE 8080

CMD ["./wallet-service"]
