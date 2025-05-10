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
  "walletId": "33333333-3333-3333-3333-333333333333",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```

## Тестирование

```bash
# Запуск интеграционных тестов
make test-repo
```
