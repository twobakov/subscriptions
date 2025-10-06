# Сервис подписок на Go + Fiber + PostgreSQL

REST API для управления подписками пользователей, Каждая подписка содержит: 
- Название сервиса
- Стоимость в рублях
- UUID пользователя
- Дату начала
- Опционально дату окончания

## Стек технологий

- Go 1.24
- [Fiber](https://github.com/gofiber/fiber) v2/v5
- PostgreSQL (Docker-контейнер)
- YAML для конфигурации
- Поддержка `.env` для задания пути к конфигу
- Swagger для документации

---

## Запуск приложения

### 1. Клонировать репозиторий

```bash
git clone https://github.com/twobakov/subscriptions.git
cd subscriptions
```

### 2. Настроить .env файл
Создайте файл .env в корне проекта с содержимым:
```env
CONFIG_PATH=config/config.yaml
```

### 3. Настроить docker-compose.yml
Запуск базы данных PostgreSQL и приложения через Docker Compose:
```bash
docker-compose up -d
```

### 4. Миграции
Миграции находятся в папке ```migrations```. Они автоматически применяются при запуске контейнера PostgreSQL через docker-compose благодаря ```volumes```:
```bash
volumes:
  - ./migrations:/docker-entrypoint-initdb.d
```

### 5. Запустить приложение
```bash
go run cmd/api/main.go
```
---

## Swagger документация
Документация к API генерируется автоматически из аннотаций с помощью `fiber-swagger` и доступна по адресу: http://localhost:8080/swagger/index.html


##  Возможности
- Добавление подписки

- Получение списка подписок

- Получить подписку по ID

- Обновить подписку

- Удалить подписку

- Посчитать суммарную стоимость подписок за определенный период

Пример запроса на создание подписки:
```json
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025"
}
```
