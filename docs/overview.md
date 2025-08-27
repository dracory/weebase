# Overview

WeeBase is a light database management module that can be embedded into any Go web application.
It provides a simple, responsive UI for managing databases, tables, and data.

Technical Details
- Built with Go >= 1.22
- Uses GORM (v2) as the ORM/DB layer
- Supports MySQL, PostgreSQL, SQLite, SQL Server database engines
- Server-rendered templates using HTML Builder (HB) with embedded CSS/JS
- Uses cookie-session with secure headers
- Uses github.com/dracory/env for configuration
- Uses github.com/dracory/api for API responses (standardized JSON envelopes)
- Uses Go standard library `log/slog` for logging
- Bootstrap 5 for responsive UI (via CDN)
- Vue 3 added for some dynamic components (via CDN); no build step required
- Bootstrap Icons for icons (via CDN)
- UI pages in `pages` directory
- API handlers in `api` directory


## Installation
- Go Module path: github.com/dracory/weebase

## Embeddable Module/SDK (Library Mode)
WeeBase can be used as an importable Go module that can be embedded into any Go web app, in addition to running as a standalone server.
The module exposes an `http.Handler` and can be mounted on a single endpoint path.
Here is a minimal embedding example:
```go
package main

import (
    "log"
    "net/http"
    weebase "github.com/dracory/weebase"
)

func main() {
    basePath := "/db"
    h := weebase.New(
        weebase.WithDrivers([]string{weebase.POSTGRES, weebase.MYSQL, weebase.SQLITE}),
        weebase.WithSafeModeDefault(true),
        weebase.WithAllowAdHocConnections(true),
        weebase.WithActionParam("action"),
        weebase.WithBasePath(basePath),
    ).Handler()

    // Mount on single endpoint; all actions via query, e.g., /db?action=browse_rows
    http.Handle(basePath, h)

    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## Standalone Server Mode
To run Weebase as a standalone server, use the `cmd/server` binary:
```sh
cd cmd/server
go run main.go
```