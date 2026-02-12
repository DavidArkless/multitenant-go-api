# Multitenant Go API

A small multi-tenant HTTP API in Go demonstrating:
- JWT authentication
- Tenant isolation via middleware
- Role-based access control (admin vs viewer)
- SQLite storage with migrations
- Automatic database seeding


## Prerequisites

- Go 1.25 or higher
- make (optional but recommended)

## Setup & Run

Clone and run:

```bash
go mod download
make run
```

Or without make:

```bash
go run cmd/api/main.go
```

The API starts on port 8080. Database migrations and seeding happen automatically on startup.

## Example Requests

Login to get a JWT token:

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@tenant1.com","password":"secret"}'
```

List projects for a tenant (requires JWT):

```bash
curl http://localhost:8080/api/v1/tenants/tenant-1/projects \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

Create a project (admin only):

```bash
curl -X POST http://localhost:8080/api/v1/tenants/tenant-1/projects \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -H "Content-Type: application/json" \
  -d '{"name":"New Project"}'
```


## API Endpoints

### Authentication

- `POST /api/v1/auth/login` - Get JWT token

### Health

- `GET /api/v1/health/liveness` - Server alive check
- `GET /api/v1/health/readiness` - Server ready check

### Tenant Resources (require authentication)

- `GET /api/v1/tenants/{tenantId}/projects` - List projects (viewer or admin)
- `GET /api/v1/tenants/{tenantId}/projects/{projectId}` - Get project (viewer or admin)
- `POST /api/v1/tenants/{tenantId}/projects` - Create project (admin only)
- `DELETE /api/v1/tenants/{tenantId}/projects/{projectId}` - Delete project (admin only)

## Seed Data (Preloaded)

The database seeds automatically on first run with two tenants and three users.

**Tenant 1** (tenant-1):
- admin@tenant1.com / secret (admin role)
- viewer@tenant1.com / secret (viewer role)
- 2 projects preloaded

**Tenant 2** (tenant-2):
- admin@tenant2.com / secret (admin role)
- 1 project preloaded

Users can only access their own tenant's data. Admins can create/delete, viewers can only read.

## Testing

Run all tests:

```bash
make test
```

Run only unit tests:

```bash
make test-unit
```

Run E2E tests:

```bash
make test-e2e
```

## Database Migrations

Migrations run automatically on startup. To manage migrations manually:

```bash
# Install migration tool
make install-migrate

# Create new migration
make migration-new name=add_users_table

# Run migrations manually
make migrate-up

# Rollback last migration
make migrate-down

# Reset database
make migrate-reset
```
