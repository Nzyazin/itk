# Wallet Service API

Сервис для управления кошельками с поддержкой высокой конкурентности.

## Функциональность

- Создание и управление кошельками
- Пополнение баланса (DEPOSIT)
- Снятие средств (WITHDRAW)
- Получение информации о балансе кошелька

## Технический стек

- Golang
- PostgreSQL
- Docker & Docker Compose

## Требования

- Go 1.20+
- Docker & Docker Compose
- PostgreSQL 14+

## Запуск приложения

```bash
# Запуск с помощью Docker Compose
docker-compose up -d
```

## API Endpoints

### Операции с кошельком

```
POST /api/v1/wallet
```

Пример запроса:
```json
{
  "walletId": "123e4567-e89b-12d3-a456-426614174000",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

### Получение баланса кошелька

```
GET /api/v1/wallets/{WALLET_UUID}
```

## Тестирование

```bash
# Запуск unit-тестов
go test ./tests/unit/...

# Запуск интеграционных тестов
go test ./tests/integration/...
```
