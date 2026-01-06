# Irrigation Analytics Service

## Overview

This repository contains the **Irrigation Analytics API**, a backend service written in Go that provides analytics and monitoring capabilities for irrigation systems in an agricultural platform.

The service is designed to be **production-oriented**, focusing on:

* Efficient SQL queries over large datasets
* Clear separation of concerns (API, data access, infrastructure)
* Observability via OpenTelemetry
* Scalability through read replicas, caching, and event-driven invalidation

---

## High-Level Architecture

```
Client
  │ HTTPS
  ▼
Nginx (TLS termination)
  │ HTTP
  ▼
Go API (Gin)
  ├─ Analytics endpoint
  ├─ Redis cache (read-through, TTL-based)
  ├─ PostgreSQL (read replica)
  ├─ OpenTelemetry SDK
  └─ Event publisher (mocked locally)

PostgreSQL (write)
  └─ Outbox-style event generation

Event System
  └─ SNS → SQS (mocked locally)

Worker Service (separate repo)
  └─ Cache invalidation + observability

OpenTelemetry Collector
  └─ Logs / Metrics / Traces export
```

Key characteristics:

* **Write and Read databases are separated** to scale analytics workloads.
* **Eventual consistency** is accepted.
* **Events are used to invalidate cache**, not to mutate read models.
* The API remains stateless and horizontally scalable.

---

## Core Feature: Irrigation Analytics

### Endpoint

```
GET /v1/farms/{farm_id}/irrigation/analytics
```

### Capabilities

* Time-series analytics (daily / weekly ISO / monthly)
* Aggregated metrics for dashboards
* Year-over-year comparisons (1 and 2 years back)
* Sector-level breakdowns
* Efficient handling of large datasets

### Default Behavior

* Timezone: **UTC**
* Default date range: **last 7 days**
* Missing farms: returns **empty metrics**, not errors

---

## Data Model

### Farms

```go
type Farm struct {
  ID   uint   `gorm:"primaryKey"`
  Name string `gorm:"not null"`
}
```

### Irrigation Sectors

```go
type IrrigationSector struct {
  ID     uint
  FarmID uint
  Name   string
}
```

### Irrigation Data

```go
type IrrigationData struct {
  ID                 uint
  FarmID             uint
  IrrigationSectorID uint
  StartTime          time.Time
  EndTime            time.Time
  NominalAmount      float32
  RealAmount         float32
  Efficiency         float32 // pre-calculated
}
```

Efficiency is calculated at write time:

```
efficiency = real_amount / nominal_amount
```

Edge cases:

* `nominal_amount = 0` → efficiency allowed (0 or NULL)
* `real_amount = NULL` → treated as 0

---

## SQL & Performance Strategy (Primary Focus)

This project prioritizes **database-level aggregation** and query efficiency.

### Query Patterns

* Time-range filtering by `farm_id` and `start_time`
* Optional filtering by `irrigation_sector_id`
* Aggregation via `date_trunc()`
* Gap handling via `generate_series()` where needed

### Indexing Strategy

Initial indexes:

```sql
CREATE INDEX idx_irrigation_farm_start_time
ON irrigation_data (farm_id, start_time);
```

Indexes are validated using:

```sql
EXPLAIN ANALYZE
```

Execution plans and optimization notes are documented in this README.

---

## Caching Strategy

* Redis is used as a **read-through cache** for analytics responses.
* Cache keys are derived from query parameters.
* **TTL-based expiration** (fixed duration).
* Cache invalidation is triggered by write-side events (best-effort).

This approach balances simplicity with performance under heavy read load.

---

## Event-Driven Design

### Why Events?

* Decouple writes from cache invalidation
* Enable eventual consistency
* Prepare the system for future extensions

### Local Development

* Events are mocked using in-memory publishers.

### Production (Planned)

* SNS for fan-out
* SQS for worker consumption
* Outbox pattern ensures reliability

The worker service lives in a **separate repository** with a clearly defined responsibility.

---

## Observability

The service is fully instrumented using **OpenTelemetry**.

### What is Collected

* HTTP request traces
* Database query spans
* Custom metrics (latency, request count, error rate)

### OpenTelemetry Collector

* Runs as a separate service (Docker Compose in dev)
* Receives OTLP data from the API
* Exports to:

  * Logging (stdout)
  * Prometheus (metrics)
  * Jaeger (traces)

This setup mirrors real production observability pipelines.

---

## Hot vs Cold Data Strategy

### Motivation

Irrigation systems can generate **high-frequency data** (e.g. one event per minute, per farm, per sector). Over time, this leads to very large tables that negatively impact:

* Query performance
* Index size and maintenance
* Cache efficiency
* Operational costs

To address this, the system adopts a **Hot / Cold data strategy**.

---

### Hot Data (PostgreSQL)

* PostgreSQL remains the **source of truth**.
* Only **recent data** is stored (default: last **30 days**, configurable via environment variables).
* All real-time analytics endpoints operate **exclusively** on hot data.
* This guarantees fast queries, small indexes, and predictable performance.

If a requested date range exceeds the hot window, the analytics endpoint **rejects the request**.

---

### Cold Data (OpenSearch)

* Historical data is archived into **OpenSearch** (or Elasticsearch-compatible engines).
* Data is stored as **daily aggregated documents per farm and sector**.
* Indices are created **per month** (e.g. `irrigation-2023-01`).
* Retention is effectively **infinite**, but configurable.

Cold data is intended for:

* Long-term historical analysis
* Exploratory analytics
* Dashboards covering years of data

A future endpoint may expose historical analytics backed by OpenSearch. This endpoint is **documented but not implemented** in this repository.

---

### Archival Process (Design Only)

A dedicated service (`irrigation-archiver`, defined in the infrastructure repository) is responsible for moving data from hot to cold storage.

**Key characteristics:**

* Runs on a **cron-based schedule**
* Processes data **by date ranges** to ensure idempotency
* Aggregates data **daily per sector** before archiving
* Deletes hot data in **small chunks** to avoid locking and long transactions

The process follows this high-level flow:

1. Select data older than the hot window cutoff
2. Aggregate daily metrics per farm and sector
3. Index aggregated documents into OpenSearch
4. Delete archived rows from PostgreSQL in chunks
5. Advance a date-based watermark to track progress

This design ensures:

* Safe retries on failure
* No duplicate logical data
* Clear operational visibility

---

### Why OpenSearch?

OpenSearch is well suited for historical irrigation analytics because it provides:

* Horizontal scalability for large volumes of time-series data
* Fast aggregations over long time ranges
* Native support for time-based indices
* Lower operational pressure on the transactional database

PostgreSQL remains optimized for **recent, high-value queries**, while OpenSearch handles **long-term analytical workloads**.

---

### Cost Considerations

This separation also enables cost optimization:

* **Hot data**: fast but expensive (PostgreSQL, Redis)
* **Cold data**: scalable and cheaper (OpenSearch)

By limiting PostgreSQL to recent data only, infrastructure costs remain predictable even as historical data grows indefinitely.

---

## Infrastructure Philosophy

### Local Development

* Docker Compose
* Separate containers for:

  * API
  * PostgreSQL (write)
  * PostgreSQL (read)
  * Redis
  * Nginx
  * OpenTelemetry Collector

### Infrastructure Repository (Future)

A separate repository defines:

* Terraform + Terragrunt structure
* AWS resources (RDS, ElastiCache, SNS, SQS, ECS)
* Networking and security boundaries

This repository intentionally contains **only structure and examples**, not a full production deployment.

---

## EC2 vs Lambda (Design Consideration)

For this service, **EC2 / container-based deployment is preferred** over Lambda.

### Rationale

* Predictable and steady API traffic
* Heavy SQL queries and connection pooling
* No cold-start risk
* Better control over:

  * DB connections
  * Caching layers
  * Observability agents

### Lambda Considerations

While Lambda offers auto-scaling and reduced ops overhead, it introduces:

* Cold start latency (even with warm strategies)
* Complex DB connection management
* Less control over long-lived processes (caches, pools)

Given the analytics-heavy nature of this service, **EC2/ECS is a better fit**.

---

## Testing Strategy

* Unit tests for analytics calculations
* Integration tests for the main endpoint (happy path)
* SQL behavior validated via `EXPLAIN ANALYZE`

---

## What Is Intentionally Out of Scope

* Authentication & authorization
* Rate limiting
* Multi-region replication
* Background recomputation of aggregates

These decisions are documented to keep the scope focused and the codebase clean.

---

## Summary

This project demonstrates:

* Strong SQL and performance-driven backend design
* Clean separation of concerns
* Event-driven thinking without over-engineering
* Realistic observability and infrastructure patterns

The implementation favors **clarity, correctness, and performance** over unnecessary complexity.
