# agg-data-per-shift

## Описание
Сервис агрегирует данные работы техники и записывает полученную информацию в БД.

## ToDo
 - [x] Подключение к RabbitMQ стриму;
 - [x] Чтение данных из стрима;
 - [x] Чтение env;
 - [ ] Обработка ошибок подключения, переподключение;

### Компоненты для запуска

Переменные окружения
```.env

# Postgres
ASD_POSTGRES_HOST="localhost"
ASD_POSTGRES_PORT="5435"
ASD_POSTGRES_DBNAME="report_bd"
SERVICE_PG_ILOGIC_USERNAME=<secret>
SERVICE_PG_ILOGIC_PASSWORD=<secret>

SERVICE_PG_NUM_PULL="10"


# RabbitMQ
ASD_RMQ_HOST="192.168.0.1"
ASD_RMQ_PORT="5432"
ASD_RMQ_VHOST="asd.asd.local.asd-test-03"
SERVICE_RMQ_ENOTIFY_USERNAME=<secret>
SERVICE_RMQ_ENOTIFY_PASSWORD=<secret>
SERVICE_RMQ_QUEUE="iLogic.Messages"
SERVICE_RMQ_NAME_CONSUMER="test_consumer"
```
