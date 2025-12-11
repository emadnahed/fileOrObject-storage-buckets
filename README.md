# File Storage Service - Production-Ready Cloud Storage System

A scalable, production-ready file storage service built with Go microservices architecture, similar to Google Drive or Dropbox. This project demonstrates cloud-native design patterns, event-driven architecture, and modern DevOps practices.

## 🎯 Project Overview

This is a comprehensive file storage backend that handles file uploads, downloads, versioning, sharing, and background processing. Built as a microservices architecture, it showcases expertise in distributed systems, cloud storage (S3), message queues (RabbitMQ), and modern observability practices.

**Status**: 🟢 Phase 1 Complete - Foundation & Infrastructure Ready

## ✨ Core Features

### ✅ Implemented (Phase 1 - Foundation)
- [x] **Project Structure** - Clean microservices architecture with shared packages
- [x] **Configuration Management** - Viper-based config with environment variable support
- [x] **Structured Logging** - Zerolog integration with middleware support
- [x] **Custom Error Handling** - Application-level errors with HTTP status mapping
- [x] **Data Models** - Complete MongoDB schemas (User, File, Folder, FileVersion)
- [x] **Docker Infrastructure** - Full docker-compose setup with all dependencies
- [x] **Database Initialization** - MongoDB with indexes and validation rules
- [x] **Development Automation** - Makefile with 25+ commands

### 🚧 Planned (Phase 2-10)
- [ ] **Pre-signed URLs** - Secure direct uploads/downloads to S3
- [ ] **Chunked Uploads** - Support for large files (>100MB) with multipart upload
- [ ] **Background Processing** - Thumbnail generation, compression, virus scanning
- [ ] **File Versioning** - Track file history with deduplication
- [ ] **Soft Deletes** - Recoverable file deletion
- [ ] **Rate Limiting** - Token bucket algorithm to prevent abuse
- [ ] **JWT Authentication** - Stateless authentication with refresh tokens
- [ ] **RBAC Authorization** - Role-based access control (user, premium, admin)
- [ ] **File Sharing** - Share files with permissions (read, write, admin)
- [ ] **Event-Driven Workers** - RabbitMQ-based async processing
- [ ] **Observability** - Prometheus metrics, Grafana dashboards, Jaeger tracing
- [ ] **API Gateway** - Single entry point with routing and circuit breakers

## 🏗️ Architecture

### Microservices (7 Core Services)

```
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway (Port 8080)                     │
│  • Single entry point for all client requests                    │
│  • Rate limiting (token bucket algorithm)                        │
│  • JWT validation and user context propagation                   │
│  • Circuit breakers for downstream service resilience            │
└───────────────────┬─────────────────────────────────────────────┘
                    │
        ┌───────────┴───────────┬───────────────┬─────────────┐
        │                       │               │             │
┌───────▼────────┐   ┌──────────▼─────┐   ┌────▼──────┐  ┌──▼─────────┐
│ Auth Service   │   │ File Service   │   │ Metadata  │  │ Versioning │
│ (Port 8081)    │   │ (Port 8082)    │   │ Service   │  │ Service    │
│                │   │                │   │ (8083)    │  │ (8085)     │
│ • User reg     │   │ • Pre-signed   │   │ • File    │  │ • Version  │
│ • JWT tokens   │   │   URLs         │   │   CRUD    │  │   history  │
│ • RBAC         │   │ • Chunked      │   │ • Search  │  │ • Restore  │
│ • Session mgmt │   │   uploads      │   │ • Folders │  │ • Dedup    │
└────────────────┘   └────────────────┘   └───────────┘  └────────────┘
                             │
                     ┌───────▼────────────┐
                     │  RabbitMQ Events   │
                     │  (Message Broker)  │
                     └───────┬────────────┘
                             │
                   ┌─────────▼──────────┬────────────────┐
           ┌───────▼────────┐  ┌────────▼────────┐  ┌───▼──────────┐
           │ Processing Svc │  │ Notification    │  │   AWS S3     │
           │ (Port 8084)    │  │ Service (8086)  │  │  (Storage)   │
           │ • Thumbnails   │  │ • Email alerts  │  │ • Files      │
           │ • Compression  │  │ • Webhooks      │  │ • Versions   │
           │ • Virus scan   │  │ • Real-time     │  │ • Thumbs     │
           └────────────────┘  └─────────────────┘  └──────────────┘
```

### Technology Stack

**Backend Framework**
- Go 1.21 - High-performance, compiled language
- Gin - Fast HTTP web framework with middleware support

**Databases & Storage**
- MongoDB - Flexible NoSQL for metadata (users, files, folders)
- AWS S3 / MinIO - Object storage for actual file content
- Redis - In-memory cache for rate limiting and sessions

**Message Queue**
- RabbitMQ - Event-driven async communication between services

**Observability**
- Prometheus - Metrics collection and alerting
- Grafana - Beautiful dashboards and visualizations
- Jaeger - Distributed tracing across microservices
- Zerolog - Structured JSON logging

**Development & Deployment**
- Docker & Docker Compose - Containerization
- Makefile - Task automation
- golangci-lint - Code quality
- Git - Version control

## 📦 Current Project Structure

```
/Users/emaad/Desktop/go_backend/
│
├── pkg/                            # ✅ Shared packages across all services
│   ├── config/                     # Configuration management with Viper
│   │   └── config.go              # Environment-based config loading
│   ├── logger/                     # Structured logging
│   │   └── logger.go              # Zerolog wrapper with Gin middleware
│   ├── errors/                     # Custom error handling
│   │   └── errors.go              # AppError type with HTTP status codes
│   └── models/                     # Data models for MongoDB
│       ├── user.go                # User, APIKey, RateLimitInfo
│       ├── file.go                # File with versioning and sharing
│       ├── folder.go              # Folder hierarchy
│       └── version.go             # FileVersion for history tracking
│
├── services/                       # 🚧 Microservices (to be implemented)
│   ├── api-gateway/               # Not yet implemented
│   ├── auth-service/              # Not yet implemented
│   ├── file-service/              # Not yet implemented
│   ├── metadata-service/          # Not yet implemented
│   ├── processing-service/        # Not yet implemented
│   ├── versioning-service/        # Not yet implemented
│   └── notification-service/      # Not yet implemented
│
├── scripts/                        # ✅ Utility scripts
│   └── init-mongo.js              # MongoDB initialization with indexes
│
├── tests/                          # 🚧 Test suites (to be implemented)
│   ├── integration/
│   └── load/
│
├── docker-compose.yml              # ✅ All infrastructure services
├── Makefile                        # ✅ Development task automation
├── .env.example                    # ✅ Configuration template
├── .gitignore                      # ✅ Git exclusions
└── go.mod                          # ✅ Go module definition
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Install Go](https://golang.org/dl/)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Make** - Usually pre-installed on macOS/Linux
- **Git** - Version control

### Installation

```bash
# 1. Clone the repository
git clone <your-repo-url>
cd go_backend

# 2. Initialize the project (creates .env, installs deps, starts Docker)
make init

# 3. Verify Docker services are running
make docker-ps

# Expected output:
# - mongodb (healthy)
# - rabbitmq (healthy)
# - minio (healthy)
# - redis (healthy)
# - prometheus (running)
# - grafana (running)
# - jaeger (running)
```

### Configuration

```bash
# 1. Copy environment template
cp .env.example .env

# 2. Edit .env with your settings
nano .env

# Important variables to configure:
# - MONGO_ROOT_PASSWORD (change from default!)
# - JWT_SECRET (use a strong random string)
# - AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY (for S3)
# - REDIS_PASSWORD (change from default!)
```

### Running Services

```bash
# Start all infrastructure services
make docker-up

# View logs from all services
make docker-logs

# View logs from specific service
make docker-logs-mongodb
make docker-logs-rabbitmq

# Stop all services
make docker-down

# Restart services
make docker-restart
```

### Accessing Services

Once `make docker-up` completes, you can access:

| Service | URL | Credentials |
|---------|-----|-------------|
| **MongoDB** | `mongodb://localhost:27017` | admin / changeme |
| **RabbitMQ Management** | http://localhost:15672 | guest / guest |
| **MinIO Console** | http://localhost:9001 | minioadmin / minioadmin |
| **Redis** | `localhost:6379` | Password: changeme |
| **Prometheus** | http://localhost:9090 | No auth |
| **Grafana** | http://localhost:3000 | admin / admin |
| **Jaeger UI** | http://localhost:16686 | No auth |

## 🛠️ Development

### Available Make Commands

```bash
make help              # Show all available commands
make install           # Install/update Go dependencies
make build             # Build all services
make test              # Run all tests
make test-unit         # Run unit tests only
make test-coverage     # Generate coverage report
make lint              # Run golangci-lint
make format            # Format code with gofmt
make docker-up         # Start Docker services
make docker-down       # Stop Docker services
make docker-logs       # View all logs
make db-shell          # Open MongoDB shell
make db-backup         # Backup MongoDB
make clean             # Clean build artifacts
make init              # First-time project setup
```

### Code Quality

```bash
# Format code
make format

# Run linters
make lint

# Run tests with coverage
make test-coverage
# Opens coverage.html in browser
```

### Database Operations

```bash
# Access MongoDB shell
make db-shell

# Common MongoDB commands:
db.users.find().pretty()           # View all users
db.files.countDocuments()          # Count files
db.files.getIndexes()              # List indexes

# Backup database
make db-backup
# Backup saved to ./backups/mongodb-backup-YYYYMMDD-HHMMSS

# Restore database
make db-restore BACKUP_DIR=./backups/mongodb-backup-20231201
```

## 📊 Data Models

### User Model

Stores user accounts with authentication, storage quotas, and rate limiting.

```go
{
  _id: ObjectId,                    // Unique user ID
  email: String (unique),           // User email (indexed)
  password_hash: String,            // Bcrypt hashed password
  name: String,                     // Display name
  role: String,                     // "user", "premium", "admin"
  storage_quota: Number,            // Max storage in bytes
  storage_used: Number,             // Current usage in bytes
  api_keys: [APIKey],               // API keys for programmatic access
  rate_limit: RateLimitInfo,        // Token bucket data
  created_at: Date,
  updated_at: Date,
  deleted_at: Date                  // Soft delete timestamp
}
```

### File Model

Stores file metadata, versioning, sharing, and upload tracking.

```go
{
  _id: ObjectId,
  user_id: ObjectId,                // Owner
  file_name: String,                // Display name
  folder_id: ObjectId,              // Parent folder
  s3_key: String (unique),          // S3 object key
  s3_bucket: String,
  file_size: Number,                // Bytes
  mime_type: String,                // Content type
  checksum: String,                 // SHA-256 hash

  // Versioning
  version: Number,                  // Current version
  is_latest: Boolean,
  parent_file_id: ObjectId,         // Previous version

  // Processing
  processing_status: String,        // "pending", "processing", "completed"
  thumbnail_url: String,
  compressed_url: String,

  // Chunked upload tracking
  upload_id: String,                // S3 multipart upload ID
  chunks: [UploadChunk],
  upload_status: String,

  // Sharing
  shared_with: [SharedUser],
  is_public: Boolean,

  created_at: Date,
  updated_at: Date,
  deleted_at: Date,
  last_accessed_at: Date
}
```

### Folder Model

Hierarchical folder structure for organizing files.

```go
{
  _id: ObjectId,
  name: String,
  user_id: ObjectId,
  parent_folder_id: ObjectId,       // null for root folders
  path: String,                     // Full path: "/Documents/Work"
  shared_with: [SharedUser],
  created_at: Date,
  updated_at: Date,
  deleted_at: Date
}
```

### FileVersion Model

Version history tracking with deduplication support.

```go
{
  _id: ObjectId,
  file_id: ObjectId,
  version_number: Number,
  s3_key: String,                   // Version storage location
  file_size: Number,
  checksum: String,                 // For deduplication
  changes_description: String,
  created_by: ObjectId,
  created_at: Date,
  restored_at: Date                 // If version was restored
}
```

## 🔐 Security Features

### Implemented
- ✅ Custom error types prevent information leakage
- ✅ Password fields excluded from JSON responses
- ✅ MongoDB validation rules on collections
- ✅ Unique constraints on sensitive fields (email, S3 keys)

### Planned
- [ ] Bcrypt password hashing (cost factor: 12)
- [ ] JWT authentication with refresh tokens
- [ ] Role-Based Access Control (RBAC)
- [ ] Rate limiting (token bucket algorithm)
- [ ] Input validation on all endpoints
- [ ] File type validation (magic numbers, not extensions)
- [ ] CORS configuration
- [ ] Secrets management (AWS Secrets Manager)
- [ ] TLS/HTTPS for all connections
- [ ] S3 bucket policies (private by default)

## 📈 Performance & Scalability

### Database Optimization
- ✅ **Compound Indexes**: (user_id, created_at), (user_id, deleted_at)
- ✅ **Unique Indexes**: email, s3_key
- ✅ **Text Indexes**: file_name (full-text search)
- ✅ **TTL Index**: activity_logs (auto-delete after 90 days)
- ✅ **Sparse Indexes**: deleted_at (only index non-null values)

### Caching Strategy (Planned)
- [ ] Redis for rate limiting state
- [ ] Redis for session storage
- [ ] Redis for frequently accessed metadata
- [ ] Pre-signed URL caching (short TTL)

### Scalability Features (Planned)
- [ ] Stateless services (horizontal scaling)
- [ ] Event-driven architecture (decouple services)
- [ ] Worker pool for concurrent processing
- [ ] S3 multipart upload (large files)
- [ ] CloudFront CDN for downloads
- [ ] MongoDB sharding (future)

## 🧪 Testing Strategy

### Planned Test Coverage
- [ ] **Unit Tests** - Business logic, utils, helpers (>80% coverage)
- [ ] **Integration Tests** - Service-to-service communication (Testcontainers)
- [ ] **API Tests** - HTTP endpoint validation (contract tests)
- [ ] **Load Tests** - Performance under load (k6 scripts)
- [ ] **End-to-End Tests** - Full user workflows

### Testing Tools
- `testify/assert` - Assertions and mocking
- `testify/mock` - Mock objects
- `testcontainers-go` - Integration test containers
- `k6` - Load testing
- `httptest` - HTTP handler testing

## 📝 API Documentation (Planned)

### Auth Service (Port 8081)
```
POST   /auth/register          Register new user
POST   /auth/login             Login and get JWT token
POST   /auth/refresh           Refresh access token
GET    /auth/me                Get current user info
```

### File Service (via API Gateway Port 8080)
```
POST   /api/v1/files/upload/init        Initiate chunked upload
POST   /api/v1/files/upload/chunk       Upload a chunk
POST   /api/v1/files/upload/complete    Complete upload
GET    /api/v1/files/:id/download       Get download URL
DELETE /api/v1/files/:id                Soft delete file
```

### Metadata Service
```
GET    /api/v1/files                    List user files
GET    /api/v1/files/:id                Get file details
PUT    /api/v1/files/:id                Update metadata
POST   /api/v1/files/:id/share          Share file
GET    /api/v1/folders                  List folders
POST   /api/v1/folders                  Create folder
```

### Versioning Service
```
GET    /api/v1/files/:id/versions       List file versions
POST   /api/v1/files/:id/versions       Create new version
POST   /api/v1/files/:id/restore/:vid   Restore version
```

## 🎓 Learning Resources

This project is designed for learning Go and cloud-native development. Each file includes extensive educational comments explaining:

### Go Concepts Covered
- Package organization and imports
- Struct definitions with tags (JSON, BSON)
- Pointers vs values
- Methods and receivers
- Interfaces (error interface)
- Error handling (no try/catch!)
- Maps, slices, and arrays
- Goroutines and channels
- Context usage
- Defer, panic, recover

### Patterns Demonstrated
- Repository pattern
- Dependency injection
- Middleware pattern
- Worker pool pattern
- Circuit breaker pattern
- Event-driven architecture
- Saga pattern
- CQRS (Command Query Responsibility Segregation)

### Technologies Explained
- **Gin** - HTTP routing, middleware, request handling
- **MongoDB** - NoSQL operations, indexing, aggregation
- **AWS S3** - Object storage, pre-signed URLs, multipart upload
- **RabbitMQ** - Message publishing, consuming, dead letter queues
- **Redis** - Caching, rate limiting, Lua scripts
- **Docker** - Containerization, networking, volumes
- **Prometheus** - Metrics collection, PromQL
- **Jaeger** - Distributed tracing, span context

## 🗺️ Roadmap

### Phase 1: Foundation ✅ **COMPLETED**
- [x] Project structure
- [x] Shared packages (config, logger, errors, models)
- [x] Docker infrastructure
- [x] MongoDB initialization
- [x] Development automation (Makefile)

### Phase 2: Authentication Service 🚧 **NEXT**
- [ ] User registration with bcrypt
- [ ] JWT token generation and validation
- [ ] Token refresh mechanism
- [ ] RBAC implementation
- [ ] User management endpoints

### Phase 3: API Gateway
- [ ] Request routing
- [ ] Rate limiting (token bucket)
- [ ] Circuit breakers
- [ ] CORS configuration
- [ ] Request ID propagation

### Phase 4: File Service
- [ ] Pre-signed URL generation
- [ ] Chunked upload implementation
- [ ] S3 integration
- [ ] RabbitMQ event publishing
- [ ] Upload/download endpoints

### Phase 5: Metadata Service
- [ ] File CRUD operations
- [ ] Folder hierarchy
- [ ] File sharing
- [ ] Search functionality
- [ ] Pagination

### Phase 6: Versioning Service
- [ ] Version tracking
- [ ] Deduplication
- [ ] Version restoration
- [ ] Retention policies

### Phase 7: Processing Service
- [ ] RabbitMQ consumers
- [ ] Image thumbnail generation
- [ ] File compression
- [ ] Worker pool implementation
- [ ] Job tracking

### Phase 8: Notification Service
- [ ] Email notifications
- [ ] Webhook delivery
- [ ] Event consumption

### Phase 9: Production Readiness
- [ ] Prometheus metrics
- [ ] Grafana dashboards
- [ ] Jaeger tracing
- [ ] Security hardening
- [ ] Comprehensive testing
- [ ] CI/CD pipeline

### Phase 10: Deployment
- [ ] Environment configuration
- [ ] Performance optimization
- [ ] Production deployment
- [ ] Monitoring and alerting
- [ ] Documentation

## 🤝 Contributing

This is a learning project. Contributions, issues, and feature requests are welcome!

### Development Workflow
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes with proper comments
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit changes: `git commit -m 'Add amazing feature'`
7. Push to branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

### Code Style
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Use `gofmt` for formatting: `make format`
- Add comprehensive comments explaining **why**, not just **what**
- Write tests for new features
- Update README.md for new features

## 📄 License

This project is for educational purposes. Feel free to use it for learning and reference.

## 🙏 Acknowledgments

- **Gin Framework** - Fast HTTP web framework
- **MongoDB** - Flexible NoSQL database
- **Zerolog** - High-performance logging
- **Docker** - Containerization platform
- **Go Community** - Amazing language and ecosystem

## 📞 Support

For questions or issues:
1. Check existing issues on GitHub
2. Review the extensive code comments
3. Consult the [official Go documentation](https://golang.org/doc/)
4. Ask in Go community forums

---

**Built with ❤️ using Go** | **Status**: Phase 1 Complete ✅ | **Next**: Auth Service 🚀
