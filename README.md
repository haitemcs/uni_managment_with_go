# University Management System (Go)

A database-focused university management REST API built with Go and PostgreSQL.

## Stack
- Go 1.23 and the standard `net/http` router
- PostgreSQL 16 with pgx
- JWT bearer authentication
- bcrypt password hashing
- Docker Compose

## Features
- Register and log in; registration creates a student account
- JWT authentication and role-based access (admin, teacher, student)
- Departments, students, courses, enrollments and grades
- SQL views, indexes, constraints, audit triggers and reporting queries
- Student weighted averages, department performance and course enrollment reports

## Run with Docker

```bash
cp .env.example .env
docker compose up --build
```

API docs are available as a simple endpoint list at [http://localhost:8080/](http://localhost:8080/). Health check: [http://localhost:8080/health](http://localhost:8080/health).

The database is exposed on host port **5433** to avoid conflicts with a local PostgreSQL service. The API container connects to PostgreSQL internally on port 5432.

## Main endpoints

| Method | Endpoint | Access |
|---|---|---|
| POST | /auth/register | Public |
| POST | /auth/login | Public |
| GET | /auth/me | Authenticated |
| GET, POST | /departments | Authenticated / admin for create |
| GET, POST | /students | Authenticated / admin for create |
| GET, POST | /courses | Authenticated / admin or teacher for create |
| POST | /enrollments | Admin or teacher |
| POST | /grades | Admin or teacher |
| GET | /reports/students/{id}/average | Authenticated |
| GET | /reports/departments | Authenticated |
| GET | /reports/courses/enrollments | Authenticated |

Requests use JSON. Send the token as `Authorization: Bearer <token>`.

### Register

```bash
curl -X POST http://localhost:8080/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"student@example.com","password":"ChangeMe123!","full_name":"Sam Student","student_number":"STU-001","department_id":1}'
```

### Login

```bash
curl -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"student@example.com","password":"ChangeMe123!"}'
```

### Promote the first account to admin (local development only)

After registering, connect to the database and update the role:

```bash
docker compose exec db psql -U university -d university_db \
  -c "UPDATE users SET role = 'admin' WHERE email = 'student@example.com';"
```

For a production system, use a controlled admin bootstrap process. Do not expose database credentials or JWT secrets.

## SQL learning path
1. Read `sql/01_schema.sql`: primary keys, foreign keys, CHECK and UNIQUE constraints, indexes.
2. Read `sql/02_analytics.sql`: views, functions and audit triggers.
3. Practice joins, subqueries, NOT EXISTS, GROUP BY, CASE and window functions in `sql/03_practice_queries.sql`.

The SQL files run automatically only when PostgreSQL creates a fresh data volume. To reset the local database and rerun them:

```bash
docker compose down -v
docker compose up --build
```

**Warning:** `docker compose down -v` deletes the project's database volume and its data.

## Project structure

```
cmd/api/                 Application entry point
internal/config/         Environment configuration
internal/database/       PostgreSQL connection pool
internal/auth/           Password hashing and JWT
internal/httpapi/        HTTP routes, middleware and handlers
sql/                     Schema, analytics and practice SQL
```

## Tests

```bash
go test ./...
```

This is an educational starter project. Before production use, add database migrations, refresh-token rotation, password reset, rate limiting, stricter CORS, structured logging, request limits and broader integration tests.
