# Base Project

A production-ready Go REST API following **Clean Architecture** with Domain-Driven Design. Provides a solid foundation for building web applications with secure authentication, user management, and session handling.

> **Note**: This is a **template project**. You should customize configurations (`docker-compose.yml`, `.env`, etc.) according to your specific requirements before deploying to production.

## Features

### Authentication & Security
- **Flexible Auth Methods**: Password-based, magic link (passwordless), or both
- **Email Verification**: Confirmation flow for new registrations
- **Password Management**: Forgot password, reset, and update flows
- **Session Management**: Secure HTTP-only cookies with multi-device support
- **Argon2 Hashing**: Industry-standard password security

### Core Capabilities
- User registration and profile management
- Session revocation (single or all sessions)
- Account status management (PENDING, ACTIVE, BLOCKED)
- Rate limiting and CORS protection
- Swagger/OpenAPI documentation

## Tech Stack

| Category | Technology |
|----------|------------|
| Language | Go 1.25 |
| Framework | Echo v4 |
| Database | PostgreSQL |
| SQL Generation | SQLC |
| Migrations | sql-migrate |
| DI Container | Uber/dig |
| Email | Resend |
| Testing | Testify + Mockery |
| Docs | Swagger |

## Project Structure

```
├── cmd/
│   ├── api/              # Application entry point
│   └── migrate/          # Migration CLI
├── internal/
│   ├── api/              # HTTP layer (handlers, middleware, routes)
│   ├── domain/           # Core entities and business rules
│   ├── service/          # Business logic layer
│   ├── infra/            # Database, external clients
│   └── mocks/            # Auto-generated test mocks
├── pkg/                  # Shared utilities
├── config/               # Configuration management
└── docs/                 # API documentation
```

## Getting Started

### Prerequisites

- Go 1.25+
- Docker and Docker Compose (recommended) or PostgreSQL installed locally
- Make

### Quick Start with Docker

1. Clone the repository:
```bash
git clone https://github.com/g-villarinho/base-api.git
cd base-project
```
```bash
cd base-project
```

2. Start PostgreSQL with Docker Compose:
```bash
docker compose up -d
```

3. Configure environment variables:
```bash
cp .env.example .env
# Edit .env if needed (default values work with docker-compose)
```

4. Install development tools and run migrations:
```bash
make setup
make migrate
```

5. Start the server:
```bash
make run
```

The API will be available at `http://localhost:5001`.

### Installation (without Docker)

1. Clone the repository:
```bash
git clone https://github.com/g-villarinho/base-api.git
```

```bash
cd base-project
```

2. Install and configure PostgreSQL, then create a database:
```bash
createdb baseproject
```

3. Install development tools:
```bash
make setup
```

4. Configure environment variables:
```bash
cp .env.example .env
# Edit .env with your PostgreSQL connection settings
```

5. Run database migrations:
```bash
make migrate
```

6. Start the server:
```bash
make run
```

The API will be available at `http://localhost:5001`.

## Development Commands

```bash
make setup          # Install dependencies (mockery, air, swag, sqlc, gotestsum)
make build          # Build binary to bin/api
make run            # Build and run server
make test           # Run all tests
make mocks          # Regenerate mocks from interfaces
make sqlc           # Regenerate SQLC code from queries
make swagger        # Generate Swagger docs
make migrate        # Apply pending migrations
make migrate-down   # Revert last migration
make migrate-status # Show migration status
```

### Running a Single Test

```bash
go test -v ./internal/service -run TestUserService_GetUser
```

## API Endpoints

### Health
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | User registration |
| POST | `/auth/login` | User login |
| DELETE | `/auth/logout` | User logout |
| GET | `/auth/verify-email` | Confirm email |
| POST | `/auth/forgot-password` | Request password reset |
| POST | `/auth/reset-password` | Reset password |
| PATCH | `/auth/password` | Update password |
| POST | `/auth/change-email/start` | Initiate email change |
| POST | `/auth/change-email/confirm` | Confirm email change |

### Magic Link (Passwordless)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register/magic-link` | Register with magic link |
| POST | `/auth/magic-link/start` | Request magic link |
| GET | `/auth/magic-link/verify` | Verify magic link |

### User (Authenticated)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/user/profile` | Get user profile |

### Sessions (Authenticated)
| Method | Endpoint | Description |
|--------|----------|-------------|
| DELETE | `/sessions/:session_id` | Revoke specific session |
| DELETE | `/sessions` | Revoke all sessions |

## Configuration

Key environment variables (see `.env.example` for complete list):

| Variable | Description | Default |
|----------|-------------|---------|
| `ENV` | Environment (development/staging/production) | development |
| `PORT` | Server port | 5001 |
| `DATABASE_DSN` | PostgreSQL connection string | - |
| `AUTH_METHOD` | Auth method (password/magic_link/both) | password |
| `SESSION_SECRET` | Secret for session tokens | - |
| `SESSION_DURATION` | Session lifetime | 168h |
| `RESEND_API_KEY` | Resend email API key | - |

## Architecture

The project follows Clean Architecture with 4 layers:

```
HTTP Request → Handler → Service → SQLC Store → PostgreSQL
                  ↓          ↓
              Middleware   Domain
```

1. **API Layer**: HTTP handlers, middleware, request/response models
2. **Service Layer**: Business logic and orchestration
3. **Domain Layer**: Core entities and domain errors
4. **Infrastructure Layer**: Database, external clients, notifications

### Key Patterns

- **Dependency Injection**: Uber/dig container for loose coupling
- **Error Mapping**: Database errors translated to domain errors
- **Type-Safe SQL**: SQLC generates Go code from SQL queries
- **Transactional Operations**: `Store.ExecTx` for atomic operations

## Testing

Tests use Testify for assertions and Mockery for auto-generated mocks:

```go
func TestUserService_GetUser(t *testing.T) {
    mockStore := new(mocks.StoreMock)
    service := service.NewUserService(mockStore)

    mockStore.On("FindUserByID", mock.Anything, userID).
        Return(&sqlc.User{...}, nil)

    user, err := service.GetUser(context.Background(), userID)

    assert.NoError(t, err)
    mockStore.AssertExpectations(t)
}
```

Regenerate mocks after interface changes:
```bash
make mocks
```

## API Documentation

Interactive API documentation is available via Swagger UI:

- **Swagger UI**: `http://localhost:5001/docs` (development only)
- **OpenAPI Spec**: `http://localhost:5001/swagger/doc.json`

All endpoints are fully documented with request/response schemas, authentication requirements, and example values.

### Additional Documentation

- Database setup guide: [docs/DATABASE.md](docs/DATABASE.md)

## License

MIT
