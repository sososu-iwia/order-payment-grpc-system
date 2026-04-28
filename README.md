# Order & Payment gRPC System

AP2 Assignment 2 — gRPC Migration & Contract-First Development

## Repository Links

| Repository | Purpose |
|---|---|
| [proto-contracts](https://github.com/sososu-iwia/proto-contracts) | Source `.proto` files. GitHub Actions generates code on push. |
| [grpc-contracts-artifacts](https://github.com/sososu-iwia/grpc-contracts-artifacts) | Auto-generated Go gRPC code. Services import this via `go get`. |
| [order-payment-grpc-system](https://github.com/sososu-iwia/order-payment-grpc-system) | Main services (this repo). |

## Architecture

```
┌─────────────┐   REST POST /orders   ┌──────────────────┐
│   Client    │ ─────────────────────► │  Order Service   │
└─────────────┘                        │  (Gin HTTP :8081) │
                                       │                  │
                                       │  gRPC Client     │
                                       │  (contracts/v1)  │
                                       └──────┬───────────┘
                                              │ gRPC ProcessPayment
                                              │ (contracts-artifacts)
                                       ┌──────▼───────────┐
                                       │ Payment Service  │
                                       │ (gRPC :50051)    │
                                       │ + Logging        │
                                       │   Interceptor    │
                                       └──────────────────┘

Streaming:
┌─────────────┐  gRPC SubscribeToOrderUpdates  ┌──────────────────┐
│ gRPC Client │ ◄──────────────────────────────│  Order Service   │
└─────────────┘    stream (DB-backed polling)   │  (gRPC :50052)   │
                                                └──────────────────┘
```

## Contract-First Flow

1. Developer edits `.proto` in `proto-contracts` repo and pushes to `main`
2. GitHub Actions runs `protoc` → generates `.pb.go` files
3. Generated files are pushed to `grpc-contracts-artifacts` with a semver tag
4. Services import the tag via `go get github.com/sososu-iwia/grpc-contracts-artifacts@v1.0.X`

## How to Run

### Prerequisites
- Docker & Docker Compose

### Start

```bash
git clone https://github.com/sososu-iwia/order-payment-grpc-system
cd order-payment-grpc-system
docker compose up --build
```

Services:
- Order Service REST: http://localhost:8085
- Payment Service REST: http://localhost:8082
- Order Service gRPC streaming: `localhost:50052`
- Payment Service gRPC: `localhost:50051`

### Environment Variables

See `.env.example`. Key variables:

| Variable | Default | Description |
|---|---|---|
| `PAYMENT_GRPC_ADDR` | `payment-service:50051` | Payment gRPC address (never hardcoded) |
| `ORDER_DB_DSN` | — | Order database connection string |
| `PAYMENT_DB_DSN` | — | Payment database connection string |
| `PORT` | `8081` / `8082` | HTTP port |
| `GRPC_PORT` | `50051` / `50052` | gRPC port |

### API Examples

**Create an order (triggers gRPC call to payment-service):**
```bash
curl -X POST http://localhost:8085/orders \
  -H "Content-Type: application/json" \
  -d '{"customer_id":"cust-1","item_name":"MacBook","amount":50000}'
```

**Subscribe to order status stream:**
```bash
grpcurl -plaintext \
  -d '{"order_id": "<YOUR_ORDER_ID>"}' \
  localhost:50052 order.OrderService/SubscribeToOrderUpdates
```

**Cancel an order:**
```bash
curl -X PATCH http://localhost:8085/orders/<id>/cancel
```

## Design Decisions

- **Clean Architecture preserved**: Domain and UseCase layers are identical to Assignment 1. Only the transport layer changed.
- **Contract-First**: Services import `grpc-contracts-artifacts` — no `.proto` files live inside the services.
- **Streaming**: `SubscribeToOrderUpdates` polls the database every second and pushes only on real status changes.
- **Error handling**: Uses `google.golang.org/grpc/status` with proper codes (InvalidArgument, NotFound, Internal, Unavailable).
- **Bonus**: Logging interceptor on Payment Service logs every RPC method name and duration.
