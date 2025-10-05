# Проект L0

Тестовое задание на Go с использованием PostgreSQL, Redis и Apache Kafka.  
Сервис получает данные заказов из Kafka, валидирует и сохраняет их в PostgreSQL с дополнительным кэшированием в Redis.  
HTTP API позволяет получать данные заказа по его идентификатору; статическая HTML‑страница (`static/index.html`) обращается к этому API и отображает информацию.

## Стек технологий

- Go
- PostgreSQL
- Redis
- Apache Kafka (+ Zookeeper)
- Docker Compose
- Swagger (OpenAPI)
- SQL миграции

## Структура проекта

Ниже приведено дерево файлов верхнего уровня и назначение основных директорий.

```
.
├── cmd/
│   └── main.go                 # Точка входа приложения (инициализация, запуск)
├── config/
│   └── config.yaml             # Основной конфигурационный файл
├── internal/
│   ├── application/            # Инициализация и связывание слоёв (bootstrapping)
│   ├── config/                 # Логика чтения/парсинга конфигурации
│   ├── messagebroker/          # Работа с Kafka (консьюмер/продьюсер)
│   ├── models/                 # Доменные структуры (заказ, элементы и т.п.)
│   ├── redis_client/           # Инициализация и функции кэша Redis
│   ├── repository/             # Репозитории работы с БД (PostgreSQL)
│   ├── router/                 # Настройка HTTP роутов и middleware
│   └── service/                # Бизнес‑логика / use-cases
├── migrations/
│   ├── 001_init_schema.up.sql  # Создание схемы БД
│   └── 001_init_schema.down.sql# Откат схемы
├── pkg/
│   ├── logger/                 # Обёртка/инициализация логгера
│   └── validator/              # Утилиты валидации входящих данных
├── docs/
│   ├── docs.go                 # Сгенерированный swagger helper (комментарии)
│   ├── swagger.yaml            # OpenAPI спецификация (YAML)
│   └── swagger.json            # OpenAPI спецификация (JSON)
├── static/
│   └── index.html              # Статическая страница (UI для просмотра заказа)
├── redisdata/
│   ├── dump.rdb                # Persist snapshot Redis (для локальной разработки)
│   └── appendonlydir/          # Директория для AOF (если включено)
├── docker-compose.yaml         # Оркестрация сервисов (DB, Kafka, Redis)
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

### Краткое описание слоёв

| Директория | Назначение |
|------------|------------|
| `cmd` | Содержит точку входа; минимальная логика, только сборка зависимостей и запуск. |
| `internal/application` | Согласованная инициализация: конфиг, логгер, брокер, репозитории, сервисы, HTTP. |
| `internal/messagebroker` | Подписка на Kafka-топик, десериализация сообщений, обработка и передача в слой сервисов. |
| `internal/models` | Доменные модели + теги сериализации/валидации. |
| `internal/repository` | SQL доступ к PostgreSQL (CRUD / выборка по ID). |
| `internal/service` | Правила бизнес-логики, координация репозиториев, кэша и брокера. |
| `internal/redis_client` | Подключение, обёртки для get/set с TTL. |
| `internal/router` | Регистрация маршрутов API и подключение Swagger/статических файлов. |
| `pkg/logger` | Универсальный логгер (уровни, формат). |
| `pkg/validator` | Повторно используемые функции валидации. |
| `migrations` | Управление схемой БД (версионирование). |
| `docs` | Swagger спецификации; генерируются и/или редактируются вручную. |
| `static` | UI для тестирования API без отдельного фронтенда. |
| `redisdata` | Том с данными Redis для локальной разработки (не нужен в продакшене). |

## Запуск проекта

Убедитесь, что установлены Docker и Docker Compose.

1. Клонирование репозитория:
   ```bash
   git clone https://github.com/DblMOKRQ/L0
   cd L0
   ```

2. Поднятие инфраструктуры:
   ```bash
   docker-compose up -d
   ```
   Запустятся: PostgreSQL, Redis, Zookeeper, Kafka.  
   При первом запуске выполните миграции (если не автоматизировано).

3. Запуск Go-приложения:
   Убедитесь, что установлен Go.
   ```bash
   go run cmd/main.go
   ```
   HTTP сервер будет доступен по адресу, указанному в `config.yaml` (по умолчанию `localhost:8080`).

4. (Опционально) Запуск тестов:
   ```bash
   go test ./...
   ```

## Конфигурация

Файл: `config/config.yaml`

Основные параметры:
- `storage`: настройки подключения к PostgreSQL (user, password, host, port, dbname, sslmode).
- `rest`: адрес HTTP сервера.
- `kafka`: брокеры и имя топика.
- `redis`: адрес, пароль и номер DB.
- `cache`: параметры TTL и лимитов.
- `log_level`: уровни `debug|info|warn|error`.

Пример (фрагмент):
```yaml
rest:
  address: "localhost:8080"
kafka:
  brokers:
    - "localhost:9092"
  topic: "orders"
```

## API Документация

После запуска доступен Swagger UI (обычно по пути `/swagger/index.html` или `/docs`, в зависимости от роутера).  
JSON/YAML спецификации лежат в `docs/`.

Основной эндпоинт получения заказа (пример):
```
GET /api/v1/orders/{id}
```

## Поток обработки данных

1. Сообщение с заказом публикуется в Kafka (формат JSON).
2. Консьюмер читает сообщение, валидирует структуру.
3. Данные сохраняются в PostgreSQL.
4. Кэширование в Redis (для ускорения последующих чтений).
5. HTTP запрос клиента:
   - Ищем в Redis; при отсутствии — берём из БД и прогреваем кэш.
6. Ответ возвращается клиенту / статической странице.

## Миграции

Для применения миграций (пример с использованием `migrate`):
```bash
migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" up
```

Откат:
```bash
migrate -path migrations -database "postgres://user:pass@localhost:5432/dbname?sslmode=disable" down 1
```

## Логирование

Уровень управляется через `log_level` в конфиге. Рекомендуется ставить `debug` на локали и `info` / `warn` в продакшене.

## Тестирование

Запуск всех тестов:
```bash
go test ./...
```

Можно отфильтровать по пакету (пример):
```bash
go test ./internal/service -run TestOrderService
```

## Рекомендации по развитию

- Добавить health-check эндпоинт (`/healthz`) для Kubernetes / Docker healthcheck.
- Ввести retry/backoff для неуспешной подачи сообщений в Kafka.
- Добавить метрики Prometheus (кол-во запросов, latency, hit ratio кэша).
- Ввести распределённый трейсинг (OpenTelemetry).
- Вынести swagger генерацию в `make generate`.

## Лицензия

См. файл [LICENSE](LICENSE).

---

Если нужны дополнительные разделы (метрики, деплой в Kubernetes, примеры сообщений Kafka) — дайте знать.
