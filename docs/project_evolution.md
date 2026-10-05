# Project Evolution

How to extend the service without breaking its structure. See [Architecture](../README.md#architecture) for the reasons behind the layout.

## The rule

Every new technology is added in three steps:

1. **Port:** if no interface fits, define one in `internal/fizzbuzz/domain/interface.go`. It speaks in domain types, never in technology types.
2. **Adapter:** implement it under `internal/fizzbuzz/infrastructure/<kind>/<technology>.go`.
3. **Wiring:** create the adapter in `cmd/server/main.go` and pass it to the service.

`domain` and `service` never import an adapter. If a change needs that, the port is missing.

## Where things go

```
internal/fizzbuzz/
├── domain/
│   └── interface.go              # ports: FizzbuzzRepository, EventPublisher, ...
├── service/                      # depends on ports only
├── handler/
│   ├── http.go                   # current REST transport
│   └── grpc_client.go            # implement client grpc
└── infrastructure/
    ├── repository/
    │   ├── inmemory_map.go       # current store
    │   ├── inmemory_heap.go      # heap data structure?
    │   ├── redis.go              # shared statistics across replicas
    │   └── postgres.go           # SQL: durable statistics
    │	└── mongodb.go            # Non-SQL: durable statistics
    ├── events/
    │   ├── kafka.go			  # implementation of publish/consumer for kafka
    │   └── rabbitmq.go			  # implementation of publish/consumer for rabbitmq
    └── grpc/
    	└── server.go 			  # implement server grpc
    └── gcp/
        ├── pubsub.go
        └── firestore.go
cmd/
├── server/main.go                # HTTP API
└── worker/main.go                # e.g. queue consumer reusing the same service
```

## Examples

### 1. Adding new statistics store (Redis, Postgres)

The port already exists: `domain.FizzbuzzRepository`. Add `infrastructure/repository/redis.go` (stubs for Redis, Postgres and a heap-based in-memory store are already there), then replace `repository.NewInmemoryRepository()` in `main.go`. No other file changes.

Redis is planned for next steps for the scalable server for enabling replicas, and then kubernetes.

### 2. Adding events and queues (Kafka, RabbitMQ)

To publish an event for each valid request:

1. Add a port in `domain`:
   ```go
   type EventPublisher interface {
       PublishRequest(ctx context.Context, params *FizzbuzzParams) error
   }
   ```
   
   
2. Implement it in `infrastructure/events/kafka.go` or `rabbitmq.go`.
3. Inject it into `service.NewService` and call it after the repository records the request.

You will have to implement the EventConsumer yourself thought.
