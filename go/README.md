# FIFA 2026 - Go Backend

**High-performance concurrent tournament simulator built with Go**

## Features
- ⚡ Concurrent simulations using goroutines
- 🔐 Real session-based login/logout, backed by a SQLite user table
- 🔒 Passwords hashed with bcrypt — nothing in plaintext, nothing hardcoded
- 🛡️ `/dashboard` and every `/api/*` route (except `/api/login`) require a valid session cookie
- 🚦 Per-IP rate limiting on `/api/login` (5 attempts/minute) to slow brute-forcing
- 📊 Monte Carlo tournament predictions
- 📝 Structured logging with slog (JSON output)
- 🎯 Elo-based win probability calculations

## Project Structure
```
go/
├── main.go              # Backend server + API endpoints
├── auth.go              # Sessions, cookies, bcrypt, rate limiter, auth middleware
├── db.go                # SQLite schema + user/session queries
├── login.html           # Login page
├── index.html           # Dashboard with simulator
├── Dockerfile            # Container build for deployment
├── .env.example          # Config template
└── go.mod                # Go module file
```

## First-Time Setup: Create a User

There are no default/demo credentials anymore — you must seed at least one user before you can log in:

```bash
cd go
SEED_PASSWORD='choose-a-strong-password' go run . -seed-admin-username=admin
```

This hashes the password with bcrypt and stores it in `users.db` (created automatically). Run it again with a different username to add more users.

## How to Run

### 1. Using `go run` (Development)
```bash
cd go
go run .
```

### 2. Using `go build` (Production)
```bash
cd go
go build -o fifa
./fifa
```

## Access the App
- **URL:** `http://localhost:8080`
- **Login Page:** `http://localhost:8080/`
- **Dashboard:** `http://localhost:8080/dashboard` (redirects to `/` if not logged in)

## Configuration

Copy `.env.example` and adjust as needed (see file for details): `ADDR`, `DB_PATH`.

## API Endpoints
- `POST /api/login` - User login (rate-limited, sets an HttpOnly session cookie)
- `POST /api/logout` - User logout (requires a valid session)
- `POST /api/simulate` - Run tournament simulation (requires a valid session)

## Logging
All events (login, logout, API calls) are logged as structured JSON to stdout.

Example log output:
```json
{"time":"2026-07-10T15:30:00.123Z","level":"INFO","msg":"Successful login","username":"admin","ip":"127.0.0.1"}
```

## Technologies
- **Language:** Go 1.21+
- **Database:** SQLite via `modernc.org/sqlite` (pure Go, no CGO — cross-compiles cleanly)
- **Hashing:** bcrypt via `golang.org/x/crypto`
- **Logging:** log/slog (built-in)
- **Frontend:** HTML5 + Vanilla JavaScript

## Deployment

Build and run the Docker image on any host with a persistent disk (Render, Fly.io, Railway, a VPS):

```bash
docker build -t fifa-go .
docker run -p 8080:8080 -v fifa_data:/data fifa-go
```

Then seed a user inside the running container (or as a one-off release/build step):
```bash
docker run --rm -v fifa_data:/data -e DB_PATH=/data/users.db -e SEED_PASSWORD='...' fifa-go -seed-admin-username=admin
```

**Important:** mount a persistent volume at `/data` — without it, the SQLite file (and every user account) is lost on every redeploy or restart. This is why serverless platforms with no persistent filesystem (e.g. Vercel/Netlify functions) are not suitable for this backend as-is.

## File Sizes & Performance
- Handles **1,000,000+ simulations** in under 2 seconds
- Uses all CPU cores for parallel processing
- Minimal memory footprint
