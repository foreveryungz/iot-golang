# Golang IoT Study Case

## 1. Project Overview

This project is a Golang backend service for an IoT-style study case. It manages master data for devices and sensors, stores monitoring data submitted through the API, and delivers that monitoring data to an external HTTP/S endpoint through a scheduled delivery process.

The main business flow is split into two stages. First, monitoring data is stored in PostgreSQL. Second, a scheduler picks eligible records and sends them through an outbound HTTP/S client. Each attempt is logged, successful deliveries are marked as `sent`, retryable failures move through `failed` and retrying states, and records that exceed the retry policy are moved to `dead_letter` for manual follow-up.

The codebase uses a layered architecture to keep responsibilities separated and testable.

```text
Request
-> Controller
-> Service
-> Repository
-> PostgreSQL
```

```text
Monitoring Data
-> Scheduler
-> Delivery Worker
-> HTTP/S Client Endpoint
-> Retry Mechanism
-> Dead Letter
```

### Key Features

- Device CRUD
- Sensor CRUD
- Monitoring Data API
- Service-level validation for monitoring data relation
- Scheduled HTTP/S Delivery
- Retry Mechanism
- Dead Letter Queue concept
- Delivery Logs
- Transactional delivery state update
- Consistent error mapping
- Layered Architecture
- Swagger API Documentation
- SQL Migration
- Unit Testing
- HTTP Client Abstraction

---

## 2. Tech Stack

| Component      | Technology     | Purpose                                       |
| -------------- | -------------- | --------------------------------------------- |
| Language       | Golang         | Main backend language                         |
| HTTP Framework | Gin            | REST API routing and request handling         |
| Database       | PostgreSQL     | Primary relational database in production     |
| ORM            | GORM           | Data access and model mapping                 |
| Migration      | golang-migrate | SQL schema versioning and migration execution |
| Scheduler      | robfig/cron    | Scheduled delivery and retry job execution    |
| Testing        | Testify        | Assertions for unit tests                     |
| Test Database  | SQLite         | In-memory database for isolated unit tests    |

| Tool             | Purpose           |
| ---------------- | ----------------- |
| Swaggo / Swagger | API Documentation |

---

## 3. Setup Instructions

### Prerequisites

- Go 1.26.2 or newer
- PostgreSQL
- `golang-migrate` CLI
- Git

### Installation

```bash
git clone <repository-url>
cd iot-golang
go mod tidy
```

### Environment Variables

```env
APP_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=iot_golang
DB_SSLMODE=disable
CLIENT_TARGET_URL=http://localhost:9000/webhook
```

Note: the current application configuration reads `CLIENT_URL` in code. For local execution, set `CLIENT_URL` to the same value as `CLIENT_TARGET_URL` unless the config key is aligned in code.

### Database Setup

1. Create a PostgreSQL database.
2. Run the SQL migrations from `database/migrations`.

```bash
migrate -path database/migrations -database "postgres://postgres:password@localhost:5432/iot_golang?sslmode=disable" up
```

### Run the Project

```bash
go run ./cmd/server
```

### Run Tests

```bash
go test ./...
```

The test suite uses SQLite in-memory databases for isolation and does not require a running PostgreSQL instance.

### Generate Swagger Docs

```bash
swag init -g cmd/server/main.go
```

---

## 4. API Documentation

Swagger UI is available at:

http://localhost:8080/swagger/index.html

### Device

| Method | Endpoint              | Description                                                          |
| ------ | --------------------- | -------------------------------------------------------------------- |
| POST   | `/api/v1/devices`     | Create a new device record.                                          |
| GET    | `/api/v1/devices`     | Retrieve all devices.                                                |
| GET    | `/api/v1/devices/:id` | Retrieve one device by ID, including related sensors when available. |
| PUT    | `/api/v1/devices/:id` | Update an existing device.                                           |
| DELETE | `/api/v1/devices/:id` | Delete a device by ID.                                               |

### Sensor

| Method | Endpoint              | Description                                                             |
| ------ | --------------------- | ----------------------------------------------------------------------- |
| POST   | `/api/v1/sensors`     | Create a new sensor linked to a device.                                 |
| GET    | `/api/v1/sensors`     | Retrieve all sensors.                                                   |
| GET    | `/api/v1/sensors/:id` | Retrieve one sensor by ID, including its related device when available. |
| PUT    | `/api/v1/sensors/:id` | Update an existing sensor.                                              |
| DELETE | `/api/v1/sensors/:id` | Delete a sensor by ID.                                                  |

### Monitoring

| Method | Endpoint                  | Description                                                    |
| ------ | ------------------------- | -------------------------------------------------------------- |
| POST   | `/api/v1/monitoring-data` | Create a monitoring data record for later delivery processing. |
| GET    | `/api/v1/monitoring-data` | Retrieve monitoring data records.                              |

### Delivery

| Method | Endpoint                          | Description                                                                  |
| ------ | --------------------------------- | ---------------------------------------------------------------------------- |
| POST   | `/api/v1/deliveries/send-pending` | Manually trigger delivery for pending monitoring data that is ready to send. |
| GET    | `/api/v1/delivery-logs`           | Retrieve recorded delivery logs for outbound attempts.                       |

---

## 5. Assumptions

- One device can have many sensors, and one sensor belongs to exactly one device.
- Inactive device and sensor status currently act as informational metadata and do not block CRUD or monitoring record creation by themselves.
- Monitoring data can only be created when the referenced device and sensor exist.
- Sensor must belong to the submitted device.
- Maximum retry count is `3`.
- Only timeout, network, transport-level errors, HTTP `429`, and HTTP `5xx` responses are retryable.
- HTTP `4xx` errors are not retried.
- Records moved to `dead_letter` require manual investigation or replay.
- Scheduler interval is configurable at the application level and can be adjusted operationally.
- Authentication and authorization are intentionally omitted because they were not requested in the study case.

---

## 6. Design Decisions

- PostgreSQL was chosen because the project stores relational operational data and benefits from strong consistency, indexing, and mature tooling.
- Layered architecture was used to separate transport logic, business rules, persistence logic, and infrastructure concerns, making the code easier to maintain and test.
- `golang-migrate` was chosen instead of `AutoMigrate` so database schema changes remain explicit, reviewable, and reproducible across environments.
- Scheduler plus retry architecture was used because outbound HTTP/S delivery is asynchronous and external endpoints may fail temporarily.
- The `dead_letter` pattern exists to stop infinite retry loops and preserve failed records for manual investigation.
- Delivery log creation and monitoring status updates are wrapped in a database transaction to prevent inconsistent delivery states.
- Service-level validation is used to ensure monitoring data references a valid device-sensor relationship.
- Controller error mapping separates validation errors, not-found errors, duplicate constraint errors, and internal server errors.
- HTTP client abstraction exists to isolate outbound delivery behavior and make the delivery service easier to test.
- Swagger documentation was added to simplify API exploration and endpoint testing during technical assessment review.
- SQLite is used for unit tests because it provides a fast, isolated in-memory database without requiring PostgreSQL during local test execution.

---

## 7. Error Handling

- Validation errors are handled at the request layer and return client-facing `400` responses when required fields are missing or malformed.
- `400 Bad Request`: invalid input or invalid device-sensor relationship.
- `404 Not Found`: device/sensor/monitoring data not found.
- `409 Conflict`: duplicate unique fields such as `device_code` or `sensor_code`.
- `500 Internal Server Error`: unexpected database or system errors.
- Database errors are returned from the repository and surfaced through the service/controller flow so persistence failures are not silently ignored.
- Retryable HTTP errors include timeout errors, network or transport failures, HTTP `429`, and HTTP `5xx` responses.
- Non-retryable HTTP errors are primarily HTTP `4xx` responses, which are treated as permanent downstream rejections.
- Retry handling updates monitoring records through `pending`, `failed`, and retrying states by incrementing `retry_count` and scheduling `next_retry_at`.
- Dead letter flow marks a record as `dead_letter` once it is no longer eligible for retry and clears retry scheduling metadata.
- Delivery logging strategy records every outbound attempt in `delivery_logs`, including attempt number, status, target URL, status code, and error message when applicable.
