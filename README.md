# Order Service

Микросервис для обработки заказов с использованием Kafka, PostgreSQL и Redis.

## Архитектура

- **Kafka** - прием заказов
- **PostgreSQL** - основное хранилище  
- **Redis** - кэш для быстрого доступа
- **HTTP API** - REST интерфейс для получения заказов

## Быстрый старт

```bash
# Запуск всех сервисов
make docker-up

# Создание Kafka топика
make create-topic

# Наполнение тестовыми данными
make seed-db

# Запуск сервиса
make run
```

Сервис будет доступен на http://localhost:8080

## Основные команды

```bash
make docker-up      # Запуск контейнеров
make docker-down    # Остановка контейнеров  
make run           # Запуск приложения
make produce-test  # Генерация тестовых заказов
make test-unit     # Запуск юнит-тестов
make test-integration # Запуск интеграционных тестов
```

## API endpoints

- `GET /api/health` - проверка здоровья
- `GET /api/order/{id}` - получение заказа по ID
- `GET /api/benchmark` - тестирование производительности
- `POST /api/orders/bulk` - массовые операции

## Тестирование

```bash
# Юнит-тесты
make test-unit

# Интеграционные тесты (требуют запущенные контейнеры)
make docker-up
make test-integration
```

## Конфигурация

Настройки через переменные окружения:
- `POSTGRES_URL` - подключение к PostgreSQL
- `REDIS_ADDR` - адрес Redis
- `KAFKA_BROKERS` - адреса Kafka брокеров
- `HTTP_ADDR` - порт HTTP сервера
