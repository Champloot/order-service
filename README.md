# Order Service

Микросервис для обработки заказов с использованием Go, Kafka, PostgreSQL и Redis.

## Архитектура

- **Kafka** - прием заказов
- **PostgreSQL** - основное хранилище  
- **Redis** - кэш для быстрого доступа
- **HTTP API** - REST интерфейс для получения заказов

## Функциональность

- 📦 Прием и обработка заказов через Kafka
- 💾 Сохранение заказов в PostgreSQL
- ⚡ Кэширование заказов в Redis для быстрого доступа
- 🔍 REST API для поиска заказов
- ✅ Валидация данных заказов
- 📊 Бенчмаркинг производительности
- 🧪 Полное покрытие тестами

## Требования

- Go 1.25+
- Docker & Docker Compose
- PostgreSQL 13+
- Redis 6+
- Kafka 7.0+

## Настройка прав Docker

**ВНИМАНИЕ**: Если вы впервые используете Docker на этом устройстве, необходимо настроить права:

```bash
# Добавляем текущего пользователя в группу docker
sudo usermod -aG docker $USER

# Применяем изменения группы (требуется перелогин или выполнение команды newgrp)
newgrp docker

# Проверяем, что Docker работает без sudo
docker ps
```
Если команда newgrp не сработала, может потребоваться:

	Выйти из системы и зайти заново

	Перезапустить терминал

	В крайнем случае - перезагрузить систему

## Старт

1. Клонирование репозитория
```bash
git clone https://github.com/Champloot/order-service
cd order-service
```

2. Настройка окружения
```bash
cp .env.example .env
```
При необходимости отредактируйте .env файл

3. Запуск инфраструктуры
```bash
make docker-up
```

4. Создание топика Kafka (в последних версиях Kafka - не требуется)
```bash
make create-topic
```

5. Заполнение базы данных тестовыми данными
Один заказ
```bash
make produce-test
```
Бэнчмарк
```bash
make seed-db
```

6. Запуск сервиса
```bash
make build
make run
```

## Использование
После запуска сервиса откройте в браузере:
```text
http://localhost:8080
```

## Генерация моков
```bash
make generate-mocks
```

## Тестирование
Unit-тесты
```bash
make test-unit
```

Integration-тесты
```bash
make test-integration
```

```bash
make docker-up			# Запуск контейнеров
make docker-down		# Остановка контейнеров  
make run				# Запуск приложения
make produce-test		# Генерация тестовых заказов
make test-unit			# Запуск юнит-тестов
make test-integration	# Запуск интеграционных тестов
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
