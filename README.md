# TaxiGo — Microservices Ride-Sharing Platform

A production-ready backend for an Uber-style ride-sharing application built with **Go**, **Kubernetes**, and modern distributed-systems patterns.

## Architecture Overview

TaxiGo follows a microservices architecture where each service has a single responsibility and communicates via **gRPC** (synchronous) and **RabbitMQ** (asynchronous events).

```
Client (React/Next.js)
        │  HTTP / WebSocket
        ▼
   API Gateway  ──gRPC──►  Driver Service
        │           └──gRPC──►  Trip Service
        │                  └──gRPC──►  Rating Service
        │
   RabbitMQ (topic exchange)
        ├── Trip events   ──►  Driver Service
        ├── Driver events ──►  Trip Service
        ├── Payment cmd   ──►  Payment Service
        └── Payment done  ──►  Rating Service
```

## Services

| Service | Port | Description |
|---|---|---|
| **api-gateway** | 8080 | Single HTTP/WebSocket entry point; routes to downstream gRPC services |
| **driver-service** | 9002 | Driver registration, availability, and trip acceptance |
| **trip-service** | 9003 | Trip lifecycle — creation, fare estimation, status updates |
| **payment-service** | 9004 | Stripe payment session creation and success handling |
| **rating-service** | 9005 | Post-trip driver ratings; triggered by payment success events |

## Tech Stack

- **Language** — Go 1.23
- **Transport** — gRPC + Protocol Buffers (service-to-service), REST + WebSocket (client-facing)
- **Messaging** — RabbitMQ with dead-letter queues and retry logic
- **Database** — MongoDB (trips), in-memory stores (drivers, ratings)
- **Observability** — Distributed tracing with OpenTelemetry → Jaeger
- **Orchestration** — Kubernetes (local: Minikube / Docker Desktop, prod: GKE)
- **Local dev** — Tilt (hot reload, live updates)
- **Payments** — Stripe Checkout Sessions

## Rating Service (new)

The `rating-service` is an original addition to the base architecture. After a payment succeeds, a `payment.event.success` message is published to RabbitMQ. The rating service consumes this event, marks the trip as eligible for rating, and exposes two gRPC endpoints:

```protobuf
service RatingService {
  rpc SubmitRating(SubmitRatingRequest)   returns (SubmitRatingResponse);
  rpc GetDriverStats(GetDriverStatsRequest) returns (GetDriverStatsResponse);
}
```

This demonstrates the **open/closed principle** at the infrastructure level — adding a new consumer to an existing event requires zero changes to upstream services.

## Installation

### Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (with Kubernetes enabled)
- [Go 1.23+](https://go.dev/dl/)
- [Tilt](https://tilt.dev/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

### Run locally

```bash
# Start the full stack (all services, RabbitMQ, Jaeger, MongoDB)
tilt up

# Monitor pods
kubectl get pods

# Open Minikube dashboard (optional)
minikube dashboard
```

### Generate proto files

```bash
make generate-proto
```

## Event Flow — Trip Scheduling

```
User requests a trip
        │
        ▼
api-gateway ──gRPC──► trip-service (CreateTrip)
        │
        ▼
trip-service publishes trip.event.created
        │
        ▼
driver-service finds available drivers ──► driver.cmd.trip_request
        │
        ▼
driver accepts ──► driver.cmd.trip_accept
        │
        ▼
trip-service assigns driver, publishes payment.cmd.create_session
        │
        ▼
payment-service creates Stripe session, publishes payment.event.session_created
        │
        ▼
User pays → payment.event.success
        │
        ├──► trip-service updates trip status
        └──► rating-service marks trip eligible for rating
```

## Production Deployment (GCP)

### 1. Build and push images

```bash
REGION=europe-west1
PROJECT_ID=<your-gcp-project-id>

for svc in api-gateway driver-service trip-service payment-service rating-service; do
  docker build \
    -t ${REGION}-docker.pkg.dev/${PROJECT_ID}/taxigo/${svc}:latest \
    --platform linux/amd64 \
    -f infra/production/docker/${svc}.Dockerfile .
  docker push ${REGION}-docker.pkg.dev/${PROJECT_ID}/taxigo/${svc}:latest
done
```

### 2. Connect to GKE and apply manifests

```bash
gcloud container clusters get-credentials taxigo --region ${REGION} --project ${PROJECT_ID}

kubectl apply -f infra/production/k8s/app-config.yaml
kubectl apply -f infra/production/k8s/secrets.yaml
kubectl apply -f infra/production/k8s/jaeger-deployment.yaml
kubectl apply -f infra/production/k8s/rabbitmq-deployment.yaml

# Wait for infra pods, then deploy services
kubectl apply -f infra/production/k8s/
```

### 3. Get the external IP

```bash
kubectl get services api-gateway
```

## Project Structure

```
TaxiGo/
├── proto/                  # .proto source files
├── shared/                 # Shared packages (messaging, tracing, env, contracts)
│   └── proto/              # Generated protobuf Go code
├── services/
│   ├── api-gateway/        # HTTP + WebSocket gateway
│   ├── driver-service/     # Driver management
│   ├── trip-service/       # Trip lifecycle
│   ├── payment-service/    # Stripe payments
│   └── rating-service/     # Post-trip driver ratings  ← added
├── infra/
│   ├── development/        # Local Docker + K8s manifests
│   └── production/         # GCP K8s manifests
└── web/                    # React/Next.js frontend
```
