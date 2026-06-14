---
name: pincode-backend-helper
description: Assists in designing, developing, optimizing, and deploying a high-performance backend API for pincode/postal code lookups, specifically tailored for fintech and customer onboarding forms. Guides the integration of Go concurrency features (such as background goroutines for CSV/data parsing and singleflight to prevent cache stampedes under high concurrent load) and caching (sync.Map, Redis) for ultra-low latency. Use this skill when the user wants to implement, extend, or deploy a pincode database, lookup endpoint, Excel/CSV importer, or optimize performance for customer address auto-fill features.
---

# Pincode Backend Helper

A skill for building, optimizing, and deploying a pincode lookup API.

## Core Architecture Guidelines

When developing the pincode service, adhere to clean architecture patterns:
1. **Model**: Defines the database schema. The pincode field MUST be indexed since lookups will be extremely frequent.
2. **Repository**: Handles GORM database operations (e.g., `GetByPincode`).
3. **Service**: Contains business logic and caching wrappers.
4. **Controller**: Handles HTTP request parsing, input validation, and JSON serialization.

---

## Technical Specifications

### 1. Database Schema Optimization
Keep the `Pincode` model struct exactly as follows (with the unique index, types, and Areas column):
```go
package models

import "time"

type Pincode struct {
	ID        uint16    `json:"id" gorm:"primaryKey"`
	Pincode   uint32    `json:"pincode" gorm:"column:pincode;uniqueIndex;not null"`
	City      string    `json:"city" gorm:"column:city;not null"`
	District  string    `json:"district" gorm:"column:district;not null"`
	State     string    `json:"state" gorm:"column:state;not null"`
	Areas     string    `json:"-" gorm:"column:areas;type:text"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;not null"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;not null"`
}
```

### 2. High-Performance Caching
Fintech forms require API responses in less than 100ms. Database queries can be slow, especially on free-tier database instances.
- **Strategy**: Implement an in-memory cache (like `go-cache` or a simple `sync.Map`) or Redis.
- **Behavior**: Check cache first. If cache miss, fetch from database, update cache, and return.
- **Example in Go**:
  ```go
  type cachedService struct {
      repo  Repository
      cache *sync.Map
  }
  ```

### 3. API Response for Fintech Onboarding
Ensure the API responds with a structured, credential-safe JSON payload containing nested areas, hiding internal IDs and database timestamps:
```json
{
  "pincode": 560001,
  "city": "Bangalore",
  "district": "Bangalore North",
  "state": "Karnataka",
  "areas": [
    "Bangalore G.P.O.",
    "Raj Bhavan"
  ]
}
```

### 4. Free-Tier Deployment & CORS
When deploying to free providers (Render, Railway, Fly.io, Hugging Face Spaces):
- **CORS**: Enable Cross-Origin Resource Sharing so frontend applications can query the API directly.
  ```go
  router.Use(cors.Default())
  ```
- **Health Check**: Add a `/health` or `/healthz` endpoint. Free tiers spin down when inactive; configure a pinging service (like UptimeRobot) to ping the health endpoint every 10–14 minutes to keep the instance active.
- **Route Restriction**: For security and resource protection, only expose the health check endpoint and the main `GET /pincode/:pincode` lookup endpoint to public online users. Restrict or disable administrative endpoints (POST, PATCH, DELETE, and bulk exports) in the routing configuration.
- **Environment Variables**: Read port and DB connection string dynamically:
  ```go
  port := os.Getenv("PORT")
  if port == "" {
      port = "8080"
  }
  ```

### 5. Memory-Safe Data Seeding (Excel/CSV Parser)
When loading a large dataset of pincodes from a CSV or Excel file:
- Do NOT read the entire file into memory at once.
- Parse the file in streams or chunks using `excelize`'s `Rows()` reader or `encoding/csv`'s line-by-line reader.
- Use database transactions in batches (e.g., 500 records per batch) to speed up insertions.

### 6. High-Concurrency & Cache Stampede Protection
Under heavy concurrent traffic, if a pincode lookup misses the cache, multiple concurrent requests might hit the database or Excel parser at the same time (Cache Stampede/Dogpiling).
- **Strategy**: Use `golang.org/x/sync/singleflight` to combine duplicate concurrent lookups into a single execution, preventing resources exhaustion.
- **Background Seeding/Parsing**: For large static assets (like 20MB+ CSV files), parse them in a background goroutine on startup into a thread-safe `sync.Map` to ensure instant `O(1)` memory lookups without blocking API boot.

### 7. Agent Context & Documentation Management (MOC Guidelines)
To prevent context loss across chat sessions, all agents MUST carefully maintain and update the `MOC.md` file in the root of the project:
- **Review MOC First**: Always read `MOC.md` first on startup to obtain a layout of all project components and understand the existing features.
- **Sync Changes**: Whenever a new file is created, modified, or deleted, or when an architecture decision is made, update `MOC.md` with appropriate wiki-links and annotations.
- **Keep Annotations Short**: Keep one-line italicized annotations up-to-date, explaining what each note/source file covers.
- **Manage Open Questions**: Append new design decisions or unresolved bottlenecks to the "Open questions" section, and remove or document them as resolved once implemented.

