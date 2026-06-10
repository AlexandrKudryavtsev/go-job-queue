# Техническое задание

## Проект

Разработать in-memory Job Queue на Go.

Система должна принимать задачи от producers, хранить их в очереди и отдавать workers для выполнения.

---

## Цель проекта

Научиться проектировать очередь задач с:

* producers;
* workers;
* ack/nack;
* retry;
* dead letter queue;
* delayed jobs;
* graceful shutdown;
* concurrency-safe storage.

---

## Основной сценарий

Producer отправляет задачу:

`POST /jobs`

Queue сохраняет задачу.

Worker получает задачу:

`GET /jobs/next`

Worker выполняет задачу и подтверждает результат:

`POST /jobs/{id}/ack`

Если задача завершилась ошибкой:

`POST /jobs/{id}/nack`

---

## Требования

### 1. HTTP API

Очередь должна запускать HTTP-сервер на порту из конфига.

Минимальные endpoints:

* `POST /jobs` — создать задачу;
* `GET /jobs/next` — получить следующую задачу;
* `POST /jobs/{id}/ack` — подтвердить успешное выполнение;
* `POST /jobs/{id}/nack` — сообщить об ошибке;
* `GET /jobs/{id}` — получить информацию о задаче;
* `GET /stats` — статистика очереди.

---

### 2. Job model

Задача должна содержать:

* id;
* type;
* payload;
* status;
* attempts;
* max_attempts;
* created_at;
* available_at;
* started_at;
* finished_at;
* last_error.

Возможные статусы:

* queued;
* processing;
* done;
* failed;
* delayed;
* dead.

---

### 3. Создание задачи

`POST /jobs`

Пример body:

```json
{
  "type": "send_email",
  "payload": {
    "to": "user@example.com",
    "subject": "Hello"
  },
  "max_attempts": 3,
  "delay_seconds": 0
}
```

Если `delay_seconds > 0`, задача не должна выдаваться worker'ам до наступления `available_at`.

---

### 4. Получение задачи worker'ом

`GET /jobs/next`

Очередь должна вернуть первую доступную задачу со статусом `queued`.

После выдачи worker'у задача должна перейти в статус `processing`.

Если доступных задач нет:

HTTP 204 No Content

---

### 5. Ack

`POST /jobs/{id}/ack`

Если задача находится в статусе `processing`, она должна перейти в статус `done`.

Необходимо заполнить `finished_at`.

Если задача не найдена:

HTTP 404 Not Found

Если статус некорректный:

HTTP 409 Conflict

---

### 6. Nack и Retry

`POST /jobs/{id}/nack`

Body:

```json
{
  "error": "smtp timeout"
}
```

Если `attempts < max_attempts`, задача должна вернуться в очередь.

Если `attempts >= max_attempts`, задача должна перейти в статус `dead`.

Для retry необходимо поддержать backoff.

Минимальный вариант:

```text
retry_delay = attempts * 5 секунд
```

---

### 7. Visibility Timeout

Если worker получил задачу, но не сделал `ack` или `nack` за заданное время, задача должна автоматически вернуться в очередь.

Пример:

```yaml
queue:
  visibility_timeout: "30s"
```

Сценарий:

```text
worker получил job
worker умер
через 30s job снова доступна
```

---

### 8. Delayed Jobs

Если задача создана с `delay_seconds`, она должна стать доступной только после `available_at`.

До этого момента она не должна выдаваться через `/jobs/next`.

---

### 9. Dead Letter Queue

Если задача исчерпала количество попыток, она должна перейти в статус `dead`.

Dead задачи должны быть доступны через endpoint:

`GET /jobs/dead`

---

### 10. Stats

`GET /stats`

Возвращает:

```json
{
  "queued": 10,
  "processing": 2,
  "done": 35,
  "dead": 1,
  "delayed": 4
}
```

---

### 11. Конфигурация

Конфигурация должна загружаться из YAML.

Пример:

```yaml
server:
  port: 8080
  read_timeout: "5s"
  write_timeout: "10s"
  shutdown_timeout: "5s"

queue:
  visibility_timeout: "30s"
  retry_base_delay: "5s"
  max_payload_size: 1048576

logger:
  format: "text"
  level: "debug"
  add_source: false
```

---

### 12. Logging

Использовать `log/slog`.

Логировать:

* старт приложения;
* остановку приложения;
* создание job;
* выдачу job worker'у;
* ack;
* nack;
* retry scheduled;
* переход job в dead letter queue;
* возврат job после visibility timeout.

---

### 13. Graceful Shutdown

При SIGINT/SIGTERM приложение должно:

* прекратить принимать новые HTTP-запросы;
* корректно остановить background checker;
* завершить текущие HTTP-запросы;
* не потерять состояние очереди в рамках текущего процесса.

---

## Нефункциональные требования

### Потокобезопасность

Очередь должна корректно работать при параллельных producers и workers.

Не допускаются:

* data race;
* выдача одной и той же job двум workers одновременно;
* потеря job при retry;
* неконсистентные статусы.

Проверка:

```bash
go test -race ./...
```

---

### Производительность

Очередь должна выдерживать:

* минимум 1000 созданий jobs;
* минимум 100 параллельных workers;
* корректную выдачу jobs без дублей.

---

### Ограничения

На первой версии не нужно делать:

* Redis;
* PostgreSQL;
* Kafka;
* persistence на диск;
* distributed locking;
* clustering.

Очередь полностью in-memory.

---

## Критерии приемки

Проект считается завершенным, если:

1. Job можно создать через `POST /jobs`.
2. Worker может получить job через `GET /jobs/next`.
3. Полученная job переходит в `processing`.
4. Job нельзя выдать двум workers одновременно.
5. `ack` переводит job в `done`.
6. `nack` возвращает job в очередь, если попытки не исчерпаны.
7. После max attempts job уходит в dead letter queue.
8. Delayed jobs не выдаются раньше `available_at`.
9. Visibility timeout возвращает зависшие jobs в очередь.
10. `/stats` показывает корректные счетчики.
11. Graceful shutdown работает.
12. `go test -race ./...` проходит без data race.

---

## Рекомендуемые этапы реализации

### Этап 1

Config, logger, HTTP server, graceful shutdown.

### Этап 2

Job model и in-memory storage.

### Этап 3

`POST /jobs` и `GET /jobs/{id}`.

### Этап 4

`GET /jobs/next` с переводом job в `processing`.

### Этап 5

`ack` и `nack`.

### Этап 6

Retry с backoff.

### Этап 7

Delayed jobs.

### Этап 8

Visibility timeout checker.

### Этап 9

Dead letter queue и `/jobs/dead`.

### Этап 10

Stats, тесты, README.
