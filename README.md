# Order Service

Connect RPC сервис для управления заказами.

## Запуск

```bash
go run .
```

Сервис запустится на порту 8081.

## API

Сервис реализует OrderService из proto-контракта:

- `CreateOrder` - создание заказа
- `GetOrder` - получение заказа по ID
- `ListOrders` - получение списка всех заказов

## Зависимости

- `github.com/demo/contracts` - proto-контракты и сгенерированный код
