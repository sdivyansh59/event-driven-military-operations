# Mission Control Command System 🎖️

A distributed event-driven military operation management system built with Go, RabbitMQ, and PostgreSQL. This system demonstrates microservices architecture, asynchronous message processing, and secure authentication with token rotation.

## 📋 Table of Contents

- [Overview](#overview)
- [System Architecture](#system-architecture)
- [Services](#services)
  - [Commander Service](#commander-service)
  - [Worker Service](#worker-service)
- [Message Queue Architecture](#message-queue-architecture)
- [Authentication & Security](#authentication--security)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [API Endpoints](#api-endpoints)
- [Message Flow](#message-flow)
- [Screenshots](#screenshots)

## 🎯 Overview

The Mission Control Command System simulates a military command and control operation where:
- **Commander's Camp (Commander Service)**: Issues mission orders and monitors mission status
- **Soldier Workers (Worker Service)**: Execute missions and report back status updates
- **Communications Hub (RabbitMQ)**: Facilitates secure, asynchronous communication between services

The system implements:
- ✅ Event-driven architecture with message queues
- ✅ Token-based authentication with automatic rotation
- ✅ Mission lifecycle management (Created → Queued → In Progress → Completed/Failed)
- ✅ Persistent storage with PostgreSQL
- ✅ Structured logging and error handling
- ✅ Dockerized deployment

## 🏗️ System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     MISSION CONTROL SYSTEM                       │
└─────────────────────────────────────────────────────────────────┘

┌──────────────────────┐           ┌──────────────────────┐
│  Commander Service   │           │   Worker Service     │
│  (Commander's Camp)  │           │  (Soldier Worker)    │
├──────────────────────┤           ├──────────────────────┤
│                      │           │                      │
│ - REST API           │           │ - Order Consumer     │
│ - Mission CRUD       │           │ - Status Producer    │
│ - Order Producer     │           │ - Token Consumer     │
│ - Status Consumer    │           │ - Mission Executor   │
│ - Token Producer     │           │                      │
│ - Auth Validator     │           │                      │
└──────────┬───────────┘           └──────────┬───────────┘
           │                                  │
           │         ┌────────────────┐       │
           └────────►│   RabbitMQ     │◄──────┘
                     │ (Message Broker)│
                     ├────────────────┤
                     │ - order_queue  │
                     │ - status_queue │
                     │ - token_queue  │
                     └────────────────┘
           │
           │
           ▼
    ┌──────────────┐
    │  PostgreSQL  │
    │  (Database)  │
    ├──────────────┤
    │ - missions   │
    └──────────────┘
```

## 🚀 Services

### Commander Service

The **Commander Service** acts as the control center for mission operations. It provides a REST API for creating and managing missions, publishes mission orders to workers, consumes status updates, and manages authentication tokens.

#### Key Components:

**1. REST API Layer** (`routes/`, `app/mission/controller.go`)
- HTTP server using Chi router
- RESTful endpoints for mission management
- Request validation and error handling

**2. Mission Management** (`app/mission/`)
- **Controller**: Handles HTTP requests and orchestrates business logic
- **Repository**: Database operations for mission persistence
- **Converter**: Transforms between entities and DTOs
- **Types**: Domain models (MissionEntity, MissionDTO)

**3. Message Producer** (`app/producer/producer.go`)
- Publishes mission orders to `order_queue`
- Publishes authentication tokens to `token_queue`
- Ensures reliable message delivery with persistent mode

**4. Status Consumer** (`app/consumer/consumer.go`)
- Consumes status updates from `status_queue`
- Validates authentication tokens
- Updates mission status in the database
- Handles token expiration and rotation

**5. Authentication** (`app/auth/`, `middleware/auth.go`)
- Token generation and validation
- Token expiration handling (30-second lifespan)
- Automatic token rotation on expiration

**6. Database Layer** (`internal-lib/database/`)
- PostgreSQL integration using Bun ORM
- Migration management
- Transaction support
- Connection pooling

#### Mission Lifecycle:

```
CREATE → QUEUED → IN_PROGRESS → COMPLETED/FAILED
   ↓        ↓           ↓              ↓
  DB     Queue      Worker          DB Update
```

#### Key Features:

- **Snowflake ID Generation**: Distributed unique ID generation
- **Structured Logging**: JSON logs with zerolog
- **Error Recovery**: Graceful error handling and retries
- **Health Checks**: Database and RabbitMQ connectivity monitoring

### Worker Service

The **Worker Service** represents field soldiers that execute missions. It consumes mission orders, simulates mission execution, reports status updates, and handles authentication token rotation.

#### Key Components:

**1. Order Consumer** (`app/consumer/consumer.go`)
- Listens to `order_queue` for new mission orders
- Acknowledges messages after processing
- QoS control for message prefetching

**2. Mission Executor**
- Simulates mission execution (random duration)
- Status transitions: QUEUED → IN_PROGRESS → COMPLETED/FAILED
- Random success/failure simulation

**3. Status Producer** (`app/producer/producer.go`)
- Publishes status updates to `status_queue`
- Includes authentication token with each message
- Ensures message delivery confirmation

**4. Token Management**
- Consumes authentication tokens from `token_queue`
- Updates token when expired
- Includes token in all status messages

**5. Logging System** (`internal-lib/utils/logger.go`)
- Logs mission execution progress
- Logs token rotation events
- Console and file-based logging

#### Worker Execution Flow:

```
1. Receive Order → 2. Update to IN_PROGRESS → 3. Execute Mission
        ↓                      ↓                        ↓
   Validate Token      Send Status Update      Simulate Work
                                                         ↓
                                    4. Send Final Status (COMPLETED/FAILED)
```

#### Key Features:

- **Automatic Token Refresh**: Handles token expiration gracefully
- **Concurrent Processing**: Goroutine-based message handling
- **Graceful Shutdown**: Proper cleanup on termination
- **Retry Mechanism**: Requeue failed messages

## 📨 Message Queue Architecture

### Queue Overview

| Queue Name      | Producer           | Consumer           | Purpose                          |
|-----------------|--------------------|--------------------|----------------------------------|
| `order_queue`   | Commander Service  | Worker Service     | Mission orders for execution     |
| `status_queue`  | Worker Service     | Commander Service  | Mission status updates           |
| `token_queue`   | Commander Service  | Worker Service     | Authentication token rotation    |

### Message Formats

**OrderMessage** (order_queue):
```json
{
  "missionID": "123456789",
  "status": "QUEUED"
}
```

**StatusMessage** (status_queue):
```json
{
  "mission_id": "123456789",
  "status": "IN_PROGRESS",
  "token": "secure-auth-token-xyz"
}
```

**TokenMessage** (token_queue):
```json
{
  "token": "new-secure-auth-token-abc",
  "expires_at": "2025-10-24T12:30:00Z"
}
```

### Queue Characteristics

- **Durability**: All queues are durable (survive broker restarts)
- **Persistence**: Messages are marked as persistent
- **Acknowledgment**: Manual acknowledgment for reliability
- **QoS**: Prefetch count of 1 for fair distribution
- **Delivery Mode**: Persistent (DeliveryMode = 2)

## 🔐 Authentication & Security

### Token-Based Authentication Flow

```
┌──────────────┐                                    ┌──────────────┐
│  Commander   │                                    │    Worker    │
│   Service    │                                    │   Service    │
└──────┬───────┘                                    └──────┬───────┘
       │                                                   │
       │ 1. Generate Initial Token                        │
       ├──────────────────────────────────────────────────►
       │   (via token_queue)                              │
       │                                                   │
       │                                                   │ 2. Store Token
       │                                                   │
       │                                                   │
       │                  3. Status Update (with token)   │
       │◄──────────────────────────────────────────────────┤
       │                                                   │
       │ 4. Validate Token                                │
       │    - If Valid: Update DB + ACK                   │
       │    - If Expired: NACK + Send New Token           │
       │                                                   │
       │ 5. New Token (if expired)                        │
       ├──────────────────────────────────────────────────►
       │   (via token_queue)                              │
       │                                                   │
       │                                                   │ 6. Update Token
       │                                                   │
       │                  7. Retry with New Token         │
       │◄──────────────────────────────────────────────────┤
       │                                                   │
```

### Token Rotation Strategy

**Event-Driven Approach:**

1. **Token Lifespan**: 30 seconds
2. **Validation**: Commander validates token on each status message
3. **Expiration Handling**:
   - Invalid/expired token → Message is NACK'd (not acknowledged)
   - Commander publishes new token to `token_queue`
   - Worker consumes new token and updates internal state
   - Message is requeued and retried with new token

4. **Logging**: All token rotations are logged in both services

### Security Benefits

- ✅ Short-lived tokens minimize exposure window
- ✅ Automatic rotation without service interruption
- ✅ Event-driven refresh mechanism
- ✅ No polling required (efficient)
- ✅ Audit trail via logging

## 🛠️ Technology Stack

### Backend
- **Language**: Go 1.21+
- **Web Framework**: Chi (HTTP router)
- **ORM**: Bun (PostgreSQL)
- **Message Broker**: RabbitMQ (AMQP 0.9.1)
- **Logging**: Zerolog
- **Dependency Injection**: Google Wire

### Infrastructure
- **Database**: PostgreSQL 16
- **Message Queue**: RabbitMQ 3.x with Management Plugin
- **Containerization**: Docker & Docker Compose

### Libraries & Tools
- `github.com/rabbitmq/amqp091-go` - RabbitMQ client
- `github.com/uptrace/bun` - SQL ORM
- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/rs/zerolog` - Structured logging
- `github.com/joho/godotenv` - Environment management
- `github.com/google/wire` - Dependency injection

## 📁 Project Structure

```
mission_control/
├── commander-service/              # Commander's Camp (Control Center)
│   ├── app/
│   │   ├── mission/               # Mission domain logic
│   │   │   ├── controller.go     # HTTP handlers
│   │   │   ├── repository.go     # Database operations
│   │   │   ├── converter.go      # DTO transformations
│   │   │   └── type.go           # Domain models
│   │   ├── consumer/              # Status message consumer
│   │   │   └── consumer.go
│   │   ├── producer/              # Order & token producer
│   │   │   └── producer.go
│   │   ├── auth/                  # Authentication logic
│   │   │   └── auth.go
│   │   ├── shared/                # Shared constants
│   │   │   └── constant.go
│   │   └── setup/                 # Application setup
│   ├── internal-lib/              # Internal libraries
│   │   ├── database/              # Database utilities
│   │   ├── snowflake/             # ID generation
│   │   └── utils/                 # Common utilities
│   ├── middleware/                # HTTP middlewares
│   ├── migrations/                # Database migrations
│   ├── routes/                    # Route definitions
│   ├── logs/                      # Application logs
│   └── main.go                    # Entry point
│
├── worker-service/                # Soldier Worker (Executor)
│   ├── app/
│   │   ├── consumer/              # Order & token consumer
│   │   │   └── consumer.go
│   │   ├── producer/              # Status producer
│   │   │   └── producer.go
│   │   ├── shared/                # Shared constants
│   │   │   └── shared.go
│   │   └── setup/                 # Application setup
│   ├── internal-lib/              # Internal libraries
│   │   ├── snowflake/             # ID utilities
│   │   └── utils/                 # Common utilities
│   ├── logs/                      # Application logs
│   └── main.go                    # Entry point
│
├── docker-compose.yaml            # Multi-service orchestration
├── screenshots/                   # System screenshots
└── README.md                      # This file
```

## 🚀 Getting Started

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for local development)
- Make (optional)

### Quick Start with Docker (Recommended)

1. **Clone the repository**
```bash
git clone <repository-url>
cd mission_control
```

2. **Create environment file** (optional - uses defaults if not provided)
```bash
cp .env.example .env
# Edit .env with your configurations if needed
```

3. **Start all services**
```bash
docker-compose up -d
```

This will:
- Build and start Commander Service (port 8080)
- Build and start Worker Service
- Start PostgreSQL (port 5432)
- Start RabbitMQ (port 5672, Management UI: 15672)
- Run database migrations automatically

4. **Access the services**
- Commander API: `http://localhost:8080`
- RabbitMQ Management: `http://localhost:15672` (admin/password)
- PostgreSQL: `localhost:5432` (postgres/postgres)

5. **View logs**
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f commander-service
docker-compose logs -f worker-service
```

6. **Stop services**
```bash
docker-compose down

# Stop and remove volumes (clean slate)
docker-compose down -v
```

### Using Pre-built Docker Images

If you want to use pre-built images from Docker Hub instead of building locally:

1. **Pull the images**
```bash
docker pull <your-dockerhub-username>/mission-control-commander:latest
docker pull <your-dockerhub-username>/mission-control-worker:latest
```

2. **Update docker-compose.yaml** to use images instead of build:
```yaml
commander-service:
  image: <your-dockerhub-username>/mission-control-commander:latest
  # Remove the 'build' section
  
worker-service:
  image: <your-dockerhub-username>/mission-control-worker:latest
  # Remove the 'build' section
```

3. **Start services**
```bash
docker-compose up -d
```

### Local Development (Without Docker)

**Prerequisites:**
- Go 1.21+
- PostgreSQL running locally
- RabbitMQ running locally

**Commander Service:**
```bash
cd commander-service

# Install dependencies
go mod download

# Run migrations
make migrate-up  # or manually run migrations

# Run service
make run  # or: go run main.go
```

**Worker Service:**
```bash
cd worker-service

# Install dependencies
go mod download

# Run service
make run  # or: go run main.go
```

### Environment Variables

Create a `.env` file in the root directory or use environment-specific configurations.

**Production (.env):**
```env
ENVIRONMENT=production
DEBUG=false
PORT=8080

# Database
POSTGRES_DB_HOST=postgres
POSTGRES_DB_PORT=5432
POSTGRES_DB_USER=postgres
POSTGRES_DB_PASSWORD=postgres
POSTGRES_DB_NAME=commanders_camp_db

# RabbitMQ
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=password
RABBITMQ_PORT=5672
RABBITMQ_MANAGEMENT_PORT=15672

# Queues
ORDER_QUEUE_NAME=order_queue
STATUS_QUEUE_NAME=status_queue
TOKEN_QUEUE_NAME=token_queue
```

**Development (.env.local):**
```env
ENVIRONMENT=development
DEBUG=true
PORT=8080

# Use localhost for local development
POSTGRES_DB_HOST=localhost
RABBITMQ_URL=amqp://admin:password@localhost:5672/
```

## 📡 API Endpoints

### Create Mission
```http
POST /missions
Content-Type: application/json

{
  "name": "Operation Eagle Eye",
  "description": "Reconnaissance mission in sector 7",
  "created_by": "Commander Alpha"
}

Response: 202 Accepted
{
  "id": "123456789",
  "name": "Operation Eagle Eye",
  "status": "QUEUED",
  "description": "Reconnaissance mission in sector 7",
  "created_by": "Commander Alpha",
  "created_at": "2025-10-24T10:00:00Z",
  "updated_at": "2025-10-24T10:00:00Z"
}
```

### Get Mission by ID
```http
GET /missions/{id}

Response: 200 OK
{
  "id": "123456789",
  "name": "Operation Eagle Eye",
  "status": "COMPLETED",
  ...
}
```

### Health Check (for both services)
```http
GET /health
1. **Start all services**
```bash
docker-compose up -d
```

2. **Create a mission via API**
```bash
curl -X POST http://localhost:8080/missions \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Operation Desert Storm",
    "description": "Secure the northern perimeter",
    "created_by": "Commander Alpha"
## 🐳 Docker Hub Deployment

### Building and Pushing Images

If you want to publish your Docker images to Docker Hub for others to use:

1. **Build the images**
```bash
# Build commander service
docker build -t <your-dockerhub-username>/mission-control-commander:latest ./commander-service

# Build worker service
docker build -t <your-dockerhub-username>/mission-control-worker:latest ./worker-service
```

2. **Tag images (optional - for versioning)**
```bash
docker tag <your-dockerhub-username>/mission-control-commander:latest \
           <your-dockerhub-username>/mission-control-commander:v1.0.0

docker tag <your-dockerhub-username>/mission-control-worker:latest \
           <your-dockerhub-username>/mission-control-worker:v1.0.0
```

3. **Login to Docker Hub**
```bash
docker login
```

4. **Push images to Docker Hub**
```bash
# Push latest tags
docker push <your-dockerhub-username>/mission-control-commander:latest
docker push <your-dockerhub-username>/mission-control-worker:latest

# Push version tags (if created)
docker push <your-dockerhub-username>/mission-control-commander:v1.0.0
docker push <your-dockerhub-username>/mission-control-worker:v1.0.0
```

### Quick Deploy Using Docker Hub Images

Create a simplified `docker-compose.prod.yaml` for users who want to use pre-built images:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: mission-control-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: commanders_camp_db
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - mission-control-network

  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: mission-control-rabbitmq
    environment:
      RABBITMQ_DEFAULT_USER: admin
      RABBITMQ_DEFAULT_PASS: password
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    networks:
      - mission-control-network

  commander-service:
    image: <your-dockerhub-username>/mission-control-commander:latest
    container_name: mission-control-commander
    environment:
      POSTGRES_DB_HOST: postgres
      POSTGRES_DB_USER: postgres
      POSTGRES_DB_PASSWORD: postgres
      POSTGRES_DB_NAME: commanders_camp_db
      RABBITMQ_URL: amqp://admin:password@rabbitmq:5672/
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - rabbitmq
    networks:
      - mission-control-network

  worker-service:
    image: <your-dockerhub-username>/mission-control-worker:latest
    container_name: mission-control-worker
    environment:
      RABBITMQ_URL: amqp://admin:password@rabbitmq:5672/
    depends_on:
      - rabbitmq
    networks:
      - mission-control-network

volumes:
  postgres_data:
  rabbitmq_data:

networks:
  mission-control-network:
    driver: bridge
```

**Usage for end users:**
```bash
# Download the compose file
curl -O https://raw.githubusercontent.com/<your-repo>/docker-compose.prod.yaml

# Start everything
docker-compose -f docker-compose.prod.yaml up -d

# Access the application
# Commander API: http://localhost:8080
# RabbitMQ UI: http://localhost:15672 (admin/password)
```

### Multi-Architecture Builds

To support multiple platforms (AMD64, ARM64):

```bash
# Create a builder
docker buildx create --use

# Build and push multi-arch images
docker buildx build --platform linux/amd64,linux/arm64 \
  -t <your-dockerhub-username>/mission-control-commander:latest \
  --push ./commander-service

docker buildx build --platform linux/amd64,linux/arm64 \
  -t <your-dockerhub-username>/mission-control-worker:latest \
  --push ./worker-service
```

## 🚀 Production Deployment Checklist

Before deploying to production:

- [ ] Update environment variables with secure passwords
- [ ] Enable HTTPS/TLS for API endpoints
- [ ] Configure proper database backups
- [ ] Set up monitoring and alerting
- [ ] Review and adjust RabbitMQ configurations
- [ ] Configure log aggregation
- [ ] Set resource limits in docker-compose
- [ ] Enable authentication for RabbitMQ management UI
- [ ] Review security best practices
- [ ] Test disaster recovery procedures

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

**Coding Standards:**
- Follow Go best practices and conventions
- Add tests for new features
- Update documentation
- Use conventional commit messages
- Ensure Docker builds succeed

## 📄 License

This project is for educational purposes.

  }'
```

3. **Get mission status**
## 🙏 Acknowledgments

- Built as a demonstration of event-driven microservices architecture
- Inspired by military command and control systems
- Thanks to the Go, RabbitMQ, and PostgreSQL communities

```bash
# Replace {mission_id} with the ID from previous response
curl http://localhost:8080/missions/{mission_id}
```

4. **Monitor RabbitMQ dashboard**
- Visit: `http://localhost:15672`
- Login: admin/password
- Navigate to Queues tab to see message flow

5. **Check logs for token rotation events**
```bash
# Commander service logs
docker-compose logs -f commander-service | grep -i token

# Worker service logs
docker-compose logs -f worker-service | grep -i token
```

6. **Verify mission status in database**
```bash
# Connect to PostgreSQL
docker exec -it military-operation-postgres psql -U postgres -d commanders_camp_db

# Query missions
SELECT id, name, status, created_at, updated_at FROM mission ORDER BY created_at DESC LIMIT 10;
```
}
```

## 🔄 Message Flow
1. Set a shorter duration for testing (modify token generation code)
### Complete Mission Execution Flow

1. **Mission Creation**
### Load Testing

Test with multiple missions simultaneously:
```bash
for i in {1..10}; do
  curl -X POST http://localhost:8080/missions \
    -H "Content-Type: application/json" \
    -d "{\"name\": \"Mission $i\", \"created_by\": \"Tester\"}" &
done
wait
```

   - User creates mission via REST API
   - Commander saves to DB (status: CREATED)
   - Commander updates status to QUEUED
   - Commander publishes to `order_queue`

2. **Worker Receives Order**
   - Worker consumes from `order_queue`
   - Worker updates status to IN_PROGRESS
   - Worker sends status update with token to `status_queue`

3. **Commander Validates Token**
   - Commander consumes from `status_queue`
   - Commander validates authentication token
   - If valid: Updates DB, sends ACK
   - If expired: Sends NACK, publishes new token to `token_queue`

4. **Token Rotation (if needed)**
   - Worker consumes new token from `token_queue`
   - Worker updates internal token state
   - Message is requeued and retried with new token

5. **Mission Execution**
   - Worker simulates mission execution (random duration)
   - Worker determines success/failure randomly

6. **Final Status Update**
   - Worker sends final status (COMPLETED/FAILED) with token
   - Commander validates, updates DB, sends ACK
   - Mission lifecycle complete

## 📸 Screenshots

### RabbitMQ Dashboard
![RabbitMQ Dashboard](screenshots/rabbitMQ-dashboard.png)
*RabbitMQ Management interface showing queue statistics*

### Message Queue Traffic
![Message Queue Traffic](screenshots/message-queue_traffic.png)
*Real-time message traffic across queues*

### Mission Database Table
![Mission Table](screenshots/mission_tbl.png)
![Mission Table 2](screenshots/mission_tbl2.png)
*PostgreSQL mission table with status tracking*

## 🔍 Key Design Decisions

### Why Event-Driven Architecture?
- **Decoupling**: Services operate independently
- **Scalability**: Easy to add more workers
- **Resilience**: Message queues buffer temporary failures
- **Asynchronous**: Non-blocking operations

### Why Token Rotation via Queue?
- **Consistency**: Same communication pattern as other messages
- **Simplicity**: No additional HTTP endpoints needed
- **Reliability**: Leverages RabbitMQ's delivery guarantees
- **Event-Driven**: Aligns with overall architecture

### Why NACK on Token Expiration?
- **Automatic Retry**: RabbitMQ requeues NACK'd messages
- **No Message Loss**: Status update will be retried with new token
- **Clear Separation**: Token validation separate from business logic

## 📝 Logging

Both services implement comprehensive logging:

**Commander Service:**
- Mission creation/updates
- Order publications
- Status consumption
- Token validation events
- Token generation and rotation

**Worker Service:**
- Order reception
- Mission execution progress
- Status publications
- Token updates
- Error conditions

Logs are written to:
- Console (stdout)
- Files (`logs/commander-service.log`, `logs/worker-service.log`)

## 🧪 Testing

### Manual Testing Flow

1. Start all services
2. Create a mission via API
3. Monitor RabbitMQ dashboard for message flow
4. Check logs for token rotation events
5. Verify mission status updates in database

### Simulating Token Expiration

The token lifespan is configurable (default 30 seconds). You can:
1. Set a shorter duration for testing
2. Monitor logs for automatic rotation
3. Verify message requeuing and retry

## 👥 Author

**Divyansh Singh**

---

**Built with ❤️ using Go, RabbitMQ, and PostgreSQL**

