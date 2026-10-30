# South Park Messaging System

A small distributed application that lets South Park characters send and receive messages asynchronously.

- **Go API**: HTTP service built with Hexagonal Architecture (Ports and Adapters). It accepts messages on `POST /messages` and publishes them to RabbitMQ.
- **RabbitMQ**: the town "mailbox". The durable queue `southpark_messages` holds messages until something consumes them.
- **Python consumer**: listens on `southpark_messages` and prints every message to the console.
- **Python producer**: a timed script that posts a random South Park quote to the API at a fixed interval.

## Architecture

```
┌────────────┐  HTTP POST   ┌──────────────────────────────┐  AMQP   ┌────────────┐  AMQP   ┌────────────┐
│  producer  │ ───────────► │            Go API            │ ──────► │  RabbitMQ  │ ──────► │  consumer  │
│ (Python)   │  /messages   │ handler → service → adapter  │         │ southpark_ │         │  (Python)  │
└────────────┘              └──────────────────────────────┘         │  messages  │         └────────────┘
                                                                     └────────────┘
```

### Go API layout

```
api/
├── app.go                                  # Composition root: loads config, builds the RabbitMQ adapter,
│                                           # injects it into the service, wires HTTP routes, graceful shutdown
├── domain/
│   └── message.go                          # Message entity + validation rules (no external dependencies)
└── internal/
    ├── common/http.go                      # JSON response helper
    └── message/
        ├── handler.go                      # Inbound (driving) adapter: HTTP → MessageService port
        ├── service.go                      # Application service; defines the outbound MessageRepository port
        └── repository/rabbit-mq.go         # Outbound (driven) adapter: publishes JSON to RabbitMQ
```

| Hexagonal concept | Where it lives |
|---|---|
| Domain | `domain.Message`, `Message.Validate()`, `domain.ErrInvalidMessage` |
| Inbound port | `MessageService` interface (`handler.go`) |
| Outbound port | `MessageRepository` interface (`service.go`) |
| Inbound adapter | `MessageHandler` (gorilla/mux HTTP handler) |
| Outbound adapter | `RabbitMQMessageRepository` (amqp091-go publisher) |
| App / wiring | `app.go`: RabbitMQ is injected into the service through the port |

The service and the handler only depend on interfaces, so the broker could be swapped out (or mocked in tests) without touching the domain or HTTP code.

## Running with Docker

### Prerequisites

- Docker with Docker Compose v2

### 1. Create the environment files

Compose loads one `.env` file per service. These files are git-ignored, so create them before by copying the .env.examples:
### 2. Start everything

```bash
docker compose up --build
```

This starts four containers:

| Service | Container | Ports |
|---|---|---|
| `rabbitmq` | `southpark-rabbitmq` | `5672` (AMQP), `15672` (management UI) |
| `api` | `southpark-api` | `8080` |
| `consumer` | `southpark-consumer` | – |
| `producer` | `southpark-producer` | – |

The API and the consumer wait for the RabbitMQ health check to pass, and both also retry the connection on startup.

### 3. Watch the messages arrive

```bash
docker compose logs -f consumer
```

```
southpark-consumer  | Waiting for messages on 'southpark_messages'...
southpark-consumer  | Eric Cartman: Respect my authoritah!
southpark-consumer  | Stan Marsh: Oh my God, they killed Kenny!
```

### 4. RabbitMQ management UI

Open http://localhost:15672 and log in with `guest` / `guest`. The `southpark_messages` queue is listed under **Queues and Streams**. To see messages pile up in the mailbox, stop the consumer (`docker compose stop consumer`) and send a few messages.

### Stop

```bash
docker compose down
```

## API

### `POST /messages`

Request body:

```json
{
  "author": "Cartman",
  "body": "Respect my authoritah!"
}
```

Both fields are required and must not be blank. Unknown fields are rejected and the body is limited to 1 MB.

Example:

```bash
curl -i -X POST http://localhost:8080/messages \
  -H "Content-Type: application/json" \
  -d '{"author": "Cartman", "body": "Respect my authoritah!"}'
```

### Responses

| Status | When | Body |
|---|---|---|
| `201 Created` | Message published to `southpark_messages` | `{"status":"message published"}` |
| `400 Bad Request` | Malformed JSON, unknown fields, or body too large | `{"error":"invalid JSON: ..."}` |
| `422 Unprocessable Entity` | `author` and/or `body` missing or blank | `{"error":"invalid message: missing required fields: author, body"}` |
| `405 Method Not Allowed` | Any method other than `POST` | – |
| `500 Internal Server Error` | Publishing to RabbitMQ failed | `{"error":"failed to publish message"}` |

Examples of failing requests:

```bash
# 400 - invalid JSON
curl -X POST http://localhost:8080/messages -H "Content-Type: application/json" -d '{"author": "Kenny"'

# 422 - missing body
curl -X POST http://localhost:8080/messages -H "Content-Type: application/json" -d '{"author": "Kenny"}'
```

## Message format on the queue

Messages are published to the default exchange with routing key `southpark_messages`. The queue is durable and the messages are persistent, so they survive a broker restart.

```json
{"body": "Respect my authoritah!", "author": "Cartman"}
```

Content type: `application/json`.

## Configuration

| Service | Variable | Default | Description |
|---|---|---|---|
| api | `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | Broker connection string |
| api | `PORT` | `8080` | HTTP listen port |
| consumer | `RABBITMQ_URL` | `amqp://guest:guest@localhost:5672/` | Broker connection string |
| producer | `API_URL` | `http://localhost:8080/messages` | Endpoint the quotes are posted to |
| producer | `INTERVAL_SECONDS` | `10` | Seconds between messages |
| producer | `QUOTES_FILE` | `producer/quotes.json` | JSON list of `{author, body}` quotes |


## Project structure

```
.
├── api/                 # Go HTTP API (hexagonal architecture)
├── client/              # Python RabbitMQ consumer
├── producer/            # Python timed script sending random quotes
├── docker/              # Dockerfiles for api, consumer, producer
└── docker-compose.yaml  # Runs the whole system
```
