# Order Service

Микросервис для обработки и управления заказами.

## О проекте

Этот проект представляет собой полноценный микросервис для обработки заказов, реализующий:

- Прием заказов через Kafka
- Хранение данных в PostgreSQL
- Кэширование в Redis для ускорения доступа
- REST API для получения информации о заказах
- Веб-интерфейс для тестирования и мониторинга
- Бенчмаркинг производительности кэша vs базы данных

## Технологии

- Backend: Go 1.19+
- Database: PostgreSQL 13
- Cache: Redis 6
- Message Broker: Kafka 7.0.0 + Zookeeper
- Containerization: Docker + Docker Compose
- HTTP Server: Native Go HTTP

## Требования

- Go 1.19 или выше
- Docker и Docker Compose
- Git

## Быстрый старт

**1. Клонирование репозитория**

```bash
git clone https://github.com/Champloot/order-service.git
cd order-service
```

**2. Запуск инфраструктуры**

```bash
make docker-up
```

*Вообще возможно понадобятся root права, и получится запустить только через sudo. Чтобы наверняка получилось лучше сделать так:*

```bash
sudo usermod -aG docker $USER
newgrp docker
```

*После этого потребуется перезагрузка и можно использовать команду для запуска инфраструктуры.*

**3. Создание топика Kafka**

```bash
make create-topic
```

*Там возможно будут выскакивать ошибки (из-за того что Кафка ещё не успел запуститься), но по итогу все отработает корректно.*

**4. Сборка и запуск сервиса**

```
make build
make run
```

Теперь сервис доступен по адресу http://localhost:8080

## Генерация тестовых данных

Для того чтобы увидеть результат работы сервиса, нужно выполнить следующие команды (например в отдельном терминале в той же директории):

```bash
make produce-test
```

*это тест демонстрирует отдельный заказ test-order-123 со всеми полями и полной информацией.*

```bash
make seed-db
```

*этот тест создает 10 заказов, созданных для оценки разницы скорости между кэшем и базой данных.*

При запуске бенчмарка, первая половина запросов берется их кэша, а втора половина из базы данных.

## API Endpoints

Получить информацию о заказе

```http
GET /api/order/{order_uid}
```

**Response**:

```json
{
  "order": {
    "order_uid": "test-order-123",
    "track_number": "WBILMTESTTRACK",
    "entry": "WBIL",
    "delivery": {
      "name": "Test Testov",
      "phone": "+9720000000",
      "zip": "2639809",
      "city": "Kiryat Mozkin",
      "address": "Ploshad Mira 15",
      "region": "Kraiot",
      "email": "test@gmail.com"
    },
    "payment": {
      "transaction": "b563feb7b2b84b6test",
      "currency": "USD",
      "provider": "wbpay",
      "amount": 1817,
      "payment_dt": 1637907727,
      "bank": "alpha",
      "delivery_cost": 1500,
      "goods_total": 317,
      "custom_fee": 0
    },
    "items": [...],
    "locale": "en",
    "customer_id": "test",
    "delivery_service": "meest",
    "date_created": "2023-01-01T00:00:00Z"
  },
  "source": "cache",
  "timing": {
    "total": "2.345ms",
    "fetch": "1.234ms",
    "source": "cache"
  }
}
```

Бенчмарк производительности

```http
GET /api/benchmark
```

**Response**:

```json
{
  "results": {
    "test-order-1": {
      "source": "cache",
      "duration": "0.456ms",
      "success": true
    }
  },
  "summary": {
    "cache_requests": 5,
    "db_requests": 5,
    "avg_cache_time": "0.523ms",
    "avg_db_time": "2.145ms",
    "speed_ratio": 4.10
  }
}
```

## Веб-интерфейс

После запуска сервиса откройте в браузере: http://localhost:8080

**Возможности веб-интерфейса**:

· Поиск заказов по ID
· Просмотр детальной информации о заказе
· Бенчмарк производительности (кэш vs база данных)
· Визуализация времени ответа
· Отслеживание источника данных (кэш/БД)

## Конфигурация

Настройки задаются через переменные окружения:

| Переменная        | По умолчанию                                                         | Описание                     |
| ----------------- | -------------------------------------------------------------------- | ---------------------------- |
| POSTGRES_CONN_STR | postgres://user:password@localhost:5432/orderservice?sslmode=disable | PostgreSQL connection string |
| REDIS_ADDR        | localhost:6379                                                       | Redis адрес                  |
| REDIS_PASSWORD    | ``                                                                   | Пароль Redis                 |
| REDIS_DB          | 0                                                                    | Redis база данных            |
| CACHE_TTL         | 24h                                                                  | Время жизни кэша             |
| KAFKA_BROKERS     | localhost:9092                                                       | Kafka брокеры                |
| KAFKA_TOPIC       | orders                                                               | Kafka топик                  |
| KAFKA_GROUP_ID    | order-service                                                        | Kafka group ID               |
| HTTP_ADDR         | :8080                                                                | HTTP порт                    |
make run
```

Теперь сервис доступен по адресу http://localhost:8080

## Тесты

Если хочется увидеть результат работы (работоспособность сервиса в принципе, бд, кэша и т.п.)

В отдельно окне терминала, в той же директории, выполните следующее:

```bash
make produce-test
```

* это тест демонстрирует отдельный заказ test-order-123

```bash
make seed-db
```

* этот тест создает 10 заказов, половина из которых берется их кэша, а втора половина из бд
* тест демонстрирует работоспособность кэша

Чтобы увидеть результаты теста 1 наберите в поле ввода заказ test-order-123

А чтобы увидеть результаты теста 2 нажмите "Run Benchmark"
