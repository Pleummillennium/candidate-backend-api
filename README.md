# Candidate Backend API

Task management system with user authentication, task cards, comments, and change logs.

## Features

- User authentication with JWT
- Task/Card management (Create, Read, Update, Delete, Archive)
- Task archiving system (Archive/Unarchive with separate views)
- Comment system with ownership validation
- Change log tracking
- Rate limiting (100 requests per minute)
- Role-based authorization
- PostgreSQL database
- Docker containerization

## Tech Stack

- **Language**: Go 1.24
- **Framework**: Gin
- **Database**: PostgreSQL 16
- **Authentication**: JWT
- **Rate Limiting**: Ulule Limiter
- **Containerization**: Docker & Docker Compose
- **API Documentation**: Swagger/OpenAPI 3.0

## Swagger API Documentation

**Interactive API documentation is available!**

```
🔗 http://localhost:8080/swagger/index.html
```

**Features:**
- 📚 Complete API documentation
- 🧪 Interactive testing interface
- 🔐 Built-in authentication testing
- 📋 Request/response examples
- 📥 Export to JSON/YAML

**Quick Start:**
1. Start the server: `docker-compose up -d` or `go run cmd/api/main.go`
2. Open browser: http://localhost:8080/swagger/index.html
3. Register a user → Copy the token
4. Click "Authorize" button → Paste token (add "Bearer " prefix)
5. Test all endpoints!

**See [SWAGGER_GUIDE.md](SWAGGER_GUIDE.md) for detailed documentation.**

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── database/             # Database connection and migrations
│   ├── handlers/             # HTTP request handlers
│   ├── middleware/           # Authentication & rate limiting
│   └── models/               # Data models
├── migrations/               # SQL migration files
├── docker/
│   ├── Dockerfile.db         # PostgreSQL Dockerfile
│   └── Dockerfile.api        # API service Dockerfile
├── docker-compose.yml        # Docker Compose configuration
├── Dockerfile                # Combined API Dockerfile
└── README.md
```

## Prerequisites

- Docker and Docker Compose installed
- OR Go 1.24+ and PostgreSQL 16+ (for local development)

## Quick Start with Docker

1. Clone the repository:
```bash
git clone <repository-url>
cd candidate-backend-api
```

2. Start the services:
```bash
docker-compose up -d
```

This will start:
- PostgreSQL database on port 5432
- API service on port 8080

3. Check if services are running:
```bash
docker-compose ps
```

4. View logs:
```bash
docker-compose logs -f api
```

5. Stop the services:
```bash
docker-compose down
```

## Local Development Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL database:
```bash
createdb candidate_db
```

3. Create `.env` file:
```bash
cp .env.example .env
```

Edit `.env` with your configuration:
```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/candidate_db?sslmode=disable
JWT_SECRET=your-secret-key-change-this-in-production
PORT=8080
```

4. Run the application:
```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

## API Endpoints

### Authentication (Public)

#### Register a new user
```
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}
```

#### Login
```
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### Tasks (Protected - Requires Authentication)

All protected endpoints require the `Authorization` header:
```
Authorization: Bearer <your-jwt-token>
```

#### Get all tasks (non-archived)
```
GET /api/tasks
```

Query Parameters (optional):
- `limit` (integer): Number of tasks per page (default: 10)
- `offset` (integer): Number of tasks to skip (default: 0)

#### Get archived tasks
```
GET /api/tasks/archived
```

Query Parameters (optional):
- `limit` (integer): Number of tasks per page (default: 10)
- `offset` (integer): Number of tasks to skip (default: 0)

#### Get a specific task
```
GET /api/tasks/:id
```

#### Create a new task
```
POST /api/tasks
Content-Type: application/json

{
  "title": "Complete project",
  "description": "Finish the backend implementation",
  "status": "To Do",
  "due_date": "2024-12-31T23:59:59Z"
}
```

Status options: `"To Do"`, `"In Progress"`, `"Done"`

#### Update a task
```
PUT /api/tasks/:id
Content-Type: application/json

{
  "title": "Updated title",
  "status": "In Progress"
}
```

Note: Only the task creator can update the task.

#### Delete a task
```
DELETE /api/tasks/:id
```

Note: Only the task creator can delete the task.

#### Archive a task
```
POST /api/tasks/:id/archive
```

Archives a task (soft delete). Only the task creator can archive the task.

#### Unarchive a task
```
POST /api/tasks/:id/unarchive
```

Restores an archived task. Only the task creator can unarchive the task.

#### Get task change logs
```
GET /api/tasks/:id/logs
```

### Comments (Protected - Requires Authentication)

#### Get all comments for a task
```
GET /api/tasks/:id/comments
```

#### Create a comment
```
POST /api/tasks/:id/comments
Content-Type: application/json

{
  "content": "This is a comment"
}
```

#### Update a comment
```
PUT /api/comments/:id
Content-Type: application/json

{
  "content": "Updated comment content"
}
```

Note: Only the comment creator can update it.

#### Delete a comment
```
DELETE /api/comments/:id
```

Note: Only the comment creator can delete it.

### Health Check
```
GET /health
```

## Authorization Rules

1. **Tasks**:
   - Any authenticated user can view all tasks (archived and non-archived)
   - Any authenticated user can create tasks
   - Only the task creator can update, delete, archive, or unarchive their tasks

2. **Comments**:
   - Any authenticated user can view comments
   - Any authenticated user can create comments
   - Only the comment creator can update or delete their comments

## Rate Limiting

- All endpoints are rate-limited to 100 requests per minute per IP address
- Rate limit headers are included in responses:
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Remaining requests
  - `X-RateLimit-Reset`: Time when the rate limit resets

## Database Schema

### Users
- id (Primary Key)
- email (Unique)
- password_hash
- name
- created_at
- updated_at

### Tasks
- id (Primary Key)
- title
- description
- status (To Do | In Progress | Done)
- creator_id (Foreign Key -> users.id)
- due_date
- archived (Boolean, default: false)
- created_at
- updated_at

### Comments
- id (Primary Key)
- task_id (Foreign Key -> tasks.id)
- user_id (Foreign Key -> users.id)
- content
- created_at
- updated_at

### Change Logs
- id (Primary Key)
- task_id (Foreign Key -> tasks.id)
- user_id (Foreign Key -> users.id)
- action
- details
- created_at

## Testing with cURL

### Register a user:
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

### Login:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Create a task (replace TOKEN with your JWT):
```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer TOKEN" \
  -d '{
    "title": "My First Task",
    "description": "Task description",
    "status": "To Do"
  }'
```

### Get all tasks (non-archived):
```bash
curl -X GET http://localhost:8080/api/tasks \
  -H "Authorization: Bearer TOKEN"
```

### Get archived tasks:
```bash
curl -X GET http://localhost:8080/api/tasks/archived \
  -H "Authorization: Bearer TOKEN"
```

### Archive a task:
```bash
curl -X POST http://localhost:8080/api/tasks/1/archive \
  -H "Authorization: Bearer TOKEN"
```

### Unarchive a task:
```bash
curl -X POST http://localhost:8080/api/tasks/1/unarchive \
  -H "Authorization: Bearer TOKEN"
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| DATABASE_URL | PostgreSQL connection string | postgres://postgres:postgres@localhost:5432/candidate_db?sslmode=disable |
| JWT_SECRET | Secret key for JWT signing | your-secret-key-change-this-in-production |
| PORT | API server port | 8080 |

## Production Deployment

1. Change the `JWT_SECRET` to a strong random value
2. Use proper PostgreSQL credentials
3. Enable SSL for database connections
4. Set up HTTPS/TLS for the API
5. Configure proper CORS settings
6. Set up log aggregation
7. Enable database backups

## License

MIT
