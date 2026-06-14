This Map of Content (MOC) serves as the developer guide and architecture map for the Go Pincode Service project. It coordinates the various application layers, database connection details, external CSV fallback logic, and load testing scripts. First-time developers can use this directory to understand the codebase structure and how requests flow through the service.

## Entry points

If you are new to the codebase, we recommend starting with these files:
- [[main.go]]
> *The main application entry point that initializes all layers, establishes the database connection, and bootstraps the Gin server.*
- [[routes/routes.go]]
> *Defines the API routes, CORS middleware policies, and endpoints exposed to clients.*
- [[services/pincode_service.go]]
> *Orchestrates the core business logic, utilizing a local cache, singleflight request deduplication, and CSV fallback resolution.*
- [[excel/pincode_excel.go]]
> *Handles parsing and searching the embedded CSV data source as a fallback for missing database entries.*

---

## Application Bootstrapping & Routing

This cluster contains files responsible for database configuration, starting the HTTP listener, and exposing endpoints.

- [[main.go]]
> *The main application entry point that initializes all layers, establishes the database connection, and bootstraps the Gin server.*
- [[connection/mysqldb.go]]
> *Implements database connection management, supporting local and remote environments via DSN string parsing.*
- [[routes/routes.go]]
> *Defines the API routes, CORS middleware policies, and endpoints exposed to clients.*

---

## Core API Logic Layer (MVC/DDD)

These files represent the vertical architecture of the application, processing requests from controller to database model.

- [[controllers/pincode_controller.go]]
> *Receives client requests, performs parameter validation, invokes services, and structures JSON responses.*
- [[services/pincode_service.go]]
> *Orchestrates the core business logic, utilizing a local cache, singleflight request deduplication, and CSV fallback resolution.*
- [[repositories/pincode_repo.go]]
> *Interfaces with GORM to perform database operations (CRUD) on the MySQL storage layer.*
- [[models/Pincode.go]]
> *Defines the Pincode database schema, GORM models, and JSON serialization tags.*

---

## Data Source Fallbacks & Performance Testing

These components enable graceful degradation via embedded files and provide tooling to test server stability.

- [[excel/pincode_excel.go]]
> *Handles parsing and searching the embedded CSV data source as a fallback for missing database entries.*
- [[scripts/loadtest.go]]
> *A high-concurrency client load testing script written in Go to verify latency, throughput, and error rates of the service.*

---

## Open questions

- **Cache Invalidation & TTL**: How does the `sync.Map` cache handle memory growth over time, and should we introduce a time-based eviction (TTL) policy or switch to a dedicated Redis instance for clustering support?
- **Excel Fallback Data Synchronization**: When a pincode is resolved from the CSV file and added to the database, is there a mechanism to prevent write conflicts if concurrent requests invoke the repository write operations?
- **Database Schema ID Limits**: The `Pincode` model uses a `uint16` data type for the ID field; since India has more than 19,000 unique pincodes, does this field have enough headroom if we ever need to support a broader schema or multiple areas separately?
- **Load Test Targets**: Should we expand the load test script to support randomized write operations (`POST` request loading) to test GORM auto-migration and insertion latency alongside read latency?

#moc #pincode-service #go-gorm #csv-processing
