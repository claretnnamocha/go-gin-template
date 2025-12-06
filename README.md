# Go Gin API

A production-ready Go web API template using the Gin framework.

## Quick Start

```bash
make up
```

## Access

- **API:** http://localhost:1954
- **Swagger:** http://localhost:1954/api/docs
- **Health:** http://localhost:1954/health
- **Prometheus:** http://localhost:9090 (with `--profile monitoring`)

## Features

- **Clean Architecture**: Layered architecture (Handler → Service → Repository)
- **Authentication**: JWT-based authentication with access and refresh tokens
- **Authorization**: Role-based access control (RBAC)
- **Validation**: Request validation with custom validators
- **Error Handling**: Global error handling with standardized responses
- **Logging**: Structured logging with Zap
- **Database**: PostgreSQL with GORM ORM
- **Caching**: Redis integration
- **Rate Limiting**: Configurable rate limiting
- **API Documentation**: Swagger/OpenAPI at `/api/docs`
- **Health Checks**: Kubernetes-ready health endpoints
- **Docker**: Multi-stage Dockerfile and docker-compose setup
- **Hot Reload**: Development mode with Air

## Commands

```bash
make help          # Show all available commands
make up            # Start all services
make down          # Stop all services
make dev           # Start with hot reload
make logs          # Follow API logs
make test          # Run tests
make swagger       # Generate Swagger docs
make migrate       # Run migrations
make seed          # Seed database
```

## Project Structure

```
.
├── cmd/
│   ├── api/
│   │   └── main.go          # Application entry point
│   └── seed/
│       └── main.go          # Database seeder
├── internal/
│   ├── dto/                 # Data Transfer Objects
│   ├── handler/             # HTTP handlers
│   ├── middleware/          # Custom middleware
│   ├── model/               # Database models
│   ├── repository/          # Data access layer
│   └── service/             # Business logic
├── pkg/
│   ├── auth/                # JWT authentication
│   ├── cache/               # Redis client
│   ├── config/              # Configuration
│   ├── database/            # Database connection
│   ├── logger/              # Logging setup
│   ├── response/            # API response helpers
│   ├── utils/               # Utilities
│   └── validator/           # Request validation
├── migrations/              # SQL migrations
├── docs/                    # Swagger documentation
├── tests/                   # Test files
├── docker-compose.yml       # Docker services
├── docker-compose.dev.yml   # Development overrides
├── Dockerfile               # Production build
├── Dockerfile.dev           # Development build
├── Makefile                 # Build commands
└── prometheus.yml           # Prometheus config
```

## API Endpoints

### Health
- `GET /health` - Health check

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh token

### Users
- `GET /api/v1/users/me` - Get current user profile
- `PUT /api/v1/users/me` - Update profile
- `PUT /api/v1/users/me/password` - Change password
- `GET /api/v1/users` - List users (admin)
- `GET /api/v1/users/:id` - Get user by ID (admin)
- `DELETE /api/v1/users/:id` - Delete user (admin)

## API Response Format

### Success Response
```json
{
  "success": true,
  "statusCode": 200,
  "message": "Operation successful",
  "data": {}
}
```

### Paginated Response
```json
{
  "success": true,
  "statusCode": 200,
  "message": "Items retrieved",
  "data": [],
  "metadata": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "totalPages": 10
  }
}
```

### Error Response
```json
{
  "success": false,
  "statusCode": 400,
  "message": "Validation failed",
  "data": {
    "code": "VALIDATION_ERROR",
    "details": [...]
  }
}
```

## Environment Variables

See `.env.example` for all configuration options.

```env
APP_PORT=1954
DB_HOST=postgres
DB_USER=golang
DB_PASSWORD=secret
DB_NAME=godb
REDIS_HOST=redis
```

## Docker Services

| Service | Port | Description |
|---------|------|-------------|
| api | 1954 | Go/Gin application |
| postgres | 5432 | PostgreSQL database |
| redis | 6379 | Redis cache |
| prometheus | 9090 | Metrics (optional) |
| pgadmin | 5050 | DB management (optional) |

## Development

```bash
# Start with hot reload
make dev

# Run tests
make test

# Generate swagger docs
make swagger

# Access database
make db-shell

# Create new migration
make migrate-create name=create_posts_table
```

## License

MIT
