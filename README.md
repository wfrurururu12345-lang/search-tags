
# Search Tags
Учебный проект для построения поисковых тегов маркетплейса на основе логов пользователей.
## Технологии
- Go 1.22
- ClickHouse 25.8
- Docker Compose
- HTML/CSS

## Запуск
1. Перед запуском поместите logs.parquet в папку data/.
2. Создать `.env` на основе `.env.example` и указать пароль ClickHouse.
3. Запустить ClickHouse:
   docker compose up -d
4. Создать таблицы и загрузить данные:
   docker compose exec -T clickhouse clickhouse-client --multiquery < sql/01_create_load.sql
5. Построить поисковые теги:
   docker compose exec -T clickhouse clickhouse-client --multiquery < sql/02_build_tags.sql
6. Запустить Go-сервис:
   go mod download
   set -a
   source .env
   set +a
   go run .

Веб-страница: http://localhost:8080
API: GET /api/tags?q=платье
SQL-загрузку логов достаточно выполнить один раз.

## Хранение данных
Используется движок MergeTree.
В таблице search_logs выбран ORDER BY (user_id, timestamp), поскольку необходимо обрабатывать последовательные поисковые запросы каждого пользователя по времени.
В таблице tags выбран ORDER BY (query, rank) для быстрого получения тегов конкретного запроса в порядке популярности.

## Допущения
- Пустые и NULL-значения исключаются из расчёта.
- При одинаковом времени запросы сортируются по тексту.
- При одинаковом числе пользователей теги сортируются по алфавиту.
- Учитываются только соседние запросы с интервалом до 30 минут.
- Для каждой пары учитываются уникальные пользователи.
- Теги строятся пакетным SQL-запросом.

## Проверки
SQL: sql/03_checks.sql
Go: go test -v ./...
Результаты находятся в go_test_results.txt и sql_check_results.txt.
