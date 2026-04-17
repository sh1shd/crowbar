# Crowbar (3X-UI) Workspace Instructions

**Project:** A web-based control panel for managing Xray-core VPN/proxy servers. Forked from 3X-UI, providing UI for configuring protocols, managing clients, monitoring traffic, and controlling Xray instances.

---

## Architecture at a Glance

```
main.go (entry point)
├── web/                 # Main web panel (Gin-based HTTP/HTTPS)
│   ├── controller/      # HTTP handlers & API routes
│   ├── service/         # Business logic layer
│   ├── job/             # Scheduled tasks (cron-based)
│   ├── websocket/       # Real-time dashboard updates
│   ├── html/            # Embedded UI templates
│   └── assets/          # CSS, JS (embedded in binary)
├── sub/                 # Subscription server (separate port, config generation)
├── xray/                # Xray-core process wrapper
├── database/            # GORM + SQLite ORM
└── util/                # Crypto, JSON, reflection, system utilities
```

**Three Core Layers:**
1. **Web Server** (`web/`) — HTTP/HTTPS panel with controllers, services, DB models, sessions, i18n
2. **Subscription Server** (`sub/`) — Separate server for generating client subscription links/configs
3. **Xray Integration** (`xray/`) — Wraps Xray-core binary, manages process, generates protocol configs

---

## Build & Run Commands

| Command | Purpose |
|---------|---------|
| `go build -ldflags "-w -s" -o build/x-ui main.go` | Build binary |
| `./x-ui` | Run with default settings (port 54321, admin:admin) |
| `./x-ui migrate` | Migrate DB from old x-ui version |
| `./x-ui setting -show` | Display current settings |
| `./x-ui setting -port 8080 -username admin -password pwd` | Update settings |
| `./x-ui cert -webCert /cert -webCertKey /key` | Set TLS certs |
| `docker-compose up` | Run via Docker (downloads from ghcr.io) |
| `go test ./...` | Run tests (minimal suite) |

**Docker Build:** Multi-stage build in Dockerfile. Requires `DockerInit.sh` to download Xray-core binary and geo-data (IP geolocation databases).

---

## Development Environment

**Requirements:**
- **Go 1.25.7+** (specified in go.mod)
- **SQLite** (embedded, no external DB)
- **Xray-core binary** — normally provided in `bin/xray-linux-amd64`; Docker build downloads it
- Optional: Docker for containerized dev

**Database:** SQLite with GORM ORM. Auto-migrates on startup. Default credentials: `admin:admin` → **always change in production**.

**Environment Variables:**
- `XUI_LOG_LEVEL` — Set log level (debug, info, notice, warning, error)
- `XUI_ENABLE_FAIL2BAN` — Enable fail2ban in Docker container
- `.env` file supported (via `godotenv`)

---

## Key Conventions & Patterns

### 1. Service Layer Pattern
Every domain has a corresponding service:
- `InboundService` — Manage inbound rules (listening ports, protocols)
- `UserService` — Manage admin users
- `SettingService` — Manage panel configuration
- `XraySettingService` — Xray-specific configs

Each service has standard CRUD methods, called by controllers.

### 2. Controller Pattern
- Inherit from `BaseController` (has user/logger context)
- Stateless Gin handlers — delegate business logic to services
- Return `SuccessMsg(data)`, `ErrorMsg(err)`, or direct `c.JSON()`
- Controllers in `web/controller/`:
  - `api.go` — General API routes
  - `inbound.go` — Inbound management
  - `xray_setting.go` — Xray-specific settings
  - `server.go` — Internal server control (reload, stop, start)

### 3. Database & Models
- GORM with SQLite backend (`database/db.go`)
- Models defined in `database/model/model.go`
- Auto-migration on startup (no manual migration scripts)
- Common pattern: `Preload("Relation")` to avoid N+1 queries

### 4. Embedded Assets
- HTML templates embedded: `//go:embed html/*`
- CSS/JS embedded: `//go:embed assets/**`
- Cannot hot-reload without rebuilding binary
- Template rendering in controllers via `c.HTML()`

### 5. WebSocket Hub
- Real-time dashboard updates via hub-based design (`web/websocket/hub.go`)
- Clients connect, hub broadcasts events (traffic updates, logs)
- Managed in `web/websocket/notifier.go`

### 6. Configuration Management
- Settings stored in SQLite (not files)
- Access via `SettingService`
- Version/name in embedded files: `config/version`, `config/name`
- CLI commands in `main.go` update DB settings

### 7. Job Scheduling (Cron)
- Located in `web/job/`
- Uses `robfig/cron/v3`
- Jobs:
  - `xray_traffic_job.go` — Query Xray traffic stats
  - `periodic_traffic_reset_job.go` — Reset traffic counters
  - `check_xray_running_job.go` — Health check
  - `clear_logs_job.go` — Clean old logs
  - `check_client_ip_job.go` — Verify client IPs

### 8. Internationalization
- English and Russian supported (`web/translation/`)
- JSON-based translation files
- Use `web/locale` package to get locale from session
- Load strings via `locale.Localize()` helper

---

## Important Gotchas & Pitfalls

| Issue | Solution |
|-------|----------|
| **Xray Binary Missing** | `bin/xray-linux-amd64` must exist. Docker build downloads via `DockerInit.sh`; for local dev, download or symlink. |
| **Admin Default Creds** | `admin:admin` is hardcoded for new DBs. **Always change in production** via `./x-ui setting` CLI. |
| **Port Conflicts** | Web server default port: 54321. Subscription server: separate port (check settings). Xray itself uses yet another port. |
| **SQLite Write Locks** | Don't run multiple instances on same DB file (they'll block). |
| **Embedded Assets Immutable** | CSS/JS changes require rebuild; not hot-reloadable in dev. Use browser DevTools for quick testing. |
| **Subscription Server Separate** | Runs on different port than web panel; clients don't access the main panel. |
| **GORM N+1 Queries** | Must explicitly `Preload()` relations; lazy loading will cause multiple queries. |
| **Xray Protocol Version Mismatch** | If Xray-core version doesn't match expectations, config generation may fail. |
| **Graceful Reload Issues** | `SIGHUP` reloads both servers but may drop active WebSocket connections. |
| **TLS Required** | Panel expects valid certs; can init with self-signed or disable (insecure). |

---

## Common Tasks & File Locations

| Task | Files |
|------|-------|
| Add new inbound protocol | `web/service/inbound.go`, `xray/config.go` |
| Add new API endpoint | `web/controller/api.go` (or new file in same dir), then register in `web/web.go` |
| Add new setting | `web/service/setting.go`, `database/model/model.go` |
| Modify UI | `web/html/` (templates), `web/assets/` (CSS/JS), rebuild required |
| Modify database schema | `database/model/model.go`; GORM auto-migrates |
| Add scheduled task | Create new file in `web/job/`, register in `web/web.go` |
| Handle Xray traffic updates | `xray/traffic.go`, `xray/client_traffic.go` |
| Fix WebSocket issues | `web/websocket/hub.go`, `web/controller/websocket.go` |
| Modify Xray config generation | `xray/config.go`, review protocol-specific code |

---

## Useful Dependencies

```go
// Web framework
"github.com/gin-gonic/gin"

// Database
"gorm.io/gorm"
"gorm.io/driver/sqlite"

// Xray protocols
"github.com/xtls/xray-core/..."

// Scheduling
"github.com/robfig/cron/v3"

// WebSocket
"github.com/gorilla/websocket"

// i18n
"github.com/nicksnyder/go-i18n/v2"

// System metrics
"github.com/shirou/gopsutil/v3"
```

---

## Tips for Productive Development

1. **Understand the flow:** Request → Controller → Service → Database/Xray. Follow this chain when debugging.
2. **Check database.db** to inspect settings, users, inbounds without UI.
3. **Use `XUI_LOG_LEVEL=debug`** environment var for verbose logs.
4. **Rebuild after HTML/CSS/JS changes**; embedded assets don't reload.
5. **Test Xray configs** with `xray test -c config.json` before committing.
6. **WebSocket debugging:** Use browser DevTools → Network → WS to inspect messages.
7. **Reset to defaults:** Delete `database.db`; next run creates fresh DB with `admin:admin`.
8. **Protocol documentation** in Xray-core repo; this panel is a UI/config wrapper.

---

## Next Steps

- Suggest **file-level instructions** (`.github/instructions/`) for specific domains (e.g., Xray config generation, WebSocket handling)
- Suggest **custom agents** if multi-step workflows are needed (e.g., "Add new protocol end-to-end")
- Suggest **hooks** to auto-format Go code on save or run tests
