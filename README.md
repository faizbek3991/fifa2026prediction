# FIFA 2026 - Tournament Prediction Engine

**Three different implementations of an interactive FIFA World Cup 2026 Monte Carlo simulator**

## 📁 Project Structure

```
FIFA2026/
├── go/                          # Go Backend + Frontend
│   ├── main.go                  # Go server: API + auth wiring
│   ├── auth.go                  # Sessions, cookies, bcrypt, rate limiter
│   ├── db.go                    # SQLite schema + queries
│   ├── login.html               # Login page
│   ├── index.html               # Dashboard with simulator
│   ├── Dockerfile, .env.example
│   ├── go.mod                   # Go module
│   └── README.md                # Go implementation guide
│
├── python/                      # Python Streamlit Backend
│   ├── app.py                   # Streamlit app with login
│   ├── auth_db.py               # SQLite schema + bcrypt user store
│   ├── seed_user.py             # CLI to create a user
│   ├── login.html               # Reference login markup (not served by app.py)
│   ├── Dockerfile, .env.example, requirements.txt
│   └── README.md                # Python implementation guide
│
├── html/                        # Standalone predictor + a small login server
│   ├── server.js                # Express server: login/session/auth-gated dashboard
│   ├── auth_db.js                # SQLite schema + bcrypt user store
│   ├── seed_user.js              # CLI to create a user
│   ├── login.html
│   ├── wc2026_interactive_predictor.html  # The simulator (served only when logged in)
│   ├── Dockerfile, .env.example, package.json
│   └── README.md                # Standalone guide
│
└── README.md                    # This file
```

---

## 🚀 Quick Start

Every implementation now requires you to create a real user first — **there are no shared demo credentials anymore**. Each backend hashes passwords with bcrypt and stores them in its own local SQLite file (git-ignored, never committed).

### **Option 1: Go Backend** (Recommended for Performance)
```bash
cd go
SEED_PASSWORD='choose-a-strong-password' go run . -seed-admin-username=admin
go run .
# Visit: http://localhost:8080
```
✅ Fast, concurrent, production-ready
✅ Real session cookies + auth middleware guarding every protected route
✅ Structured JSON logging
✅ Goroutines for parallel processing

---

### **Option 2: Python Streamlit** (Recommended for UI)
```bash
cd python
pip install -r requirements.txt
python seed_user.py --username admin
streamlit run app.py
# Visit: http://localhost:8501
```
✅ Beautiful interactive interface
✅ Session-based authentication backed by SQLite + bcrypt
✅ Real-time charts and visualizations
✅ Easy to customize

---

### **Option 3: Standalone HTML + small login server**
```bash
cd html
npm install
npm run seed-user -- --username admin
npm start
# Visit: http://localhost:3000
```
✅ Same Monte Carlo simulator, still runs client-side
✅ Real server-side login gate (the old client-side-only login has been removed — it stored "auth" in `localStorage` and was trivially bypassable)
✅ Pure HTML/CSS/JavaScript for the simulator itself

---

## ⚙️ Features Comparison

| Feature | Go | Python | Standalone |
|---------|----|---------|----|
| Login System | ✅ (sessions + bcrypt) | ✅ (session_state + bcrypt) | ✅ (Express server + bcrypt) |
| Logout Button | ✅ | ✅ | ✅ |
| Logging | ✅ (JSON) | ✅ | ✅ (JSON) |
| Performance | ⚡⚡⚡ | ⚡⚡ | ⚡⚡ |
| Interactive Charts | ✅ | ✅✅ | ✅ |
| Concurrent Processing | ✅ (Goroutines) | ❌ | ❌ |
| Installation Required | Go 1.21+ | Python 3.8+ | Node 18+ |
| Browser Dependency | ✅ | ✅ | ✅ |
| Learning Value | High | Medium | Medium |

---

## 📊 What Each Implementation Does

### Go Implementation
- **Backend:** Lightweight HTTP server with goroutines for parallel simulations
- **Authentication:** Form-based login/logout with credential validation
- **API:** RESTful endpoints for login, logout, and simulation
- **Logging:** Structured JSON logging to stdout
- **Frontend:** Modern HTML/CSS with vanilla JavaScript

### Python Implementation
- **Backend:** Streamlit framework for rapid UI development
- **Authentication:** Session-based login using Streamlit's session_state
- **Interface:** Sidebar controls for easy parameter adjustment
- **Visualization:** Interactive Plotly charts
- **Logging:** Python's logging module for event tracking

### Standalone Implementation
- **Simulation:** Still 100% client-side Monte Carlo simulation once logged in
- **Backend:** Small Express + SQLite server whose only job is the login gate
- **UI:** Custom styled components with smooth animations
- **Performance:** No network latency for the simulation itself - instant results

---

## 🎯 Tournament Simulation Logic

All three implementations use the same core algorithm:

1. **Elo Rating System** - Teams ranked by rating
2. **Win Probability** - Calculated using Elo formula: `P(A beats B) = 1 / (1 + 10^((B-A)/400))`
3. **Monte Carlo Method** - Simulate tournament thousands of times
4. **Results Analysis** - Aggregate outcomes to show probabilities

### Tournament Bracket (Quarterfinal Stage)
```
Quarter-finals:
  France      vs  Morocco      → Winner to SF1
  Spain       vs  Belgium      → Winner to SF1
  Norway      vs  England      → Winner to SF2
  Argentina   vs  Switzerland  → Winner to SF2

Semi-finals:
  SF1 Winner  vs  SF1 Winner   → Winner to Final
  SF2 Winner  vs  SF2 Winner   → Winner to Final

Final:
  Finalist 1  vs  Finalist 2   → Champion
```

---

## 🛠️ Technology Stack

### Go Implementation
- **Language:** Go 1.21+
- **HTTP Server:** net/http (built-in)
- **Logging:** log/slog (built-in)
- **Concurrency:** Goroutines & sync.WaitGroup
- **Frontend:** HTML5, CSS3, Vanilla JavaScript

### Python Implementation
- **Framework:** Streamlit
- **Data:** Pandas
- **Visualization:** Plotly
- **Logging:** Python logging module
- **Requirements:** streamlit, pandas, plotly

### Standalone Implementation
- **Frontend:** HTML5, CSS3, JavaScript, Chart.js
- **Backend:** Node.js + Express + better-sqlite3 + bcryptjs (login only)
- **Styling:** Custom CSS with design variables

---

## 📝 Logging

### Go (Structured JSON)
```json
{"time":"2026-07-10T15:30:00.123Z","level":"INFO","msg":"User logged in successfully","username":"admin","ip":"127.0.0.1"}
```

### Python (Standard Logging)
```
2026-07-10 15:30:00,123 - INFO - Login attempt - Username: admin, Success: True
```

---

## 🎓 Learning Resources

Each implementation demonstrates different concepts:

**Go Implementation:**
- Goroutines and concurrent programming
- RESTful API design
- Session management
- Structured logging
- Production-ready code patterns

**Python Implementation:**
- Rapid app development with Streamlit
- Session state management
- Data visualization with Plotly
- Interactive UI components
- Data processing with Pandas

**Standalone Implementation:**
- Client-side simulation
- DOM manipulation
- Event handling
- Chart rendering
- Pure JavaScript algorithms

---

## 📈 Performance Benchmarks

Running 100,000 tournament simulations:

| Implementation | Time | CPU Cores |
|---|---|---|
| **Go** | ~100-150ms | Uses all cores |
| **Python** | ~500-800ms | Single core |
| **Standalone** | ~1000-1500ms | Single core |

*Results may vary based on hardware*

---

## 🚀 Deployment

All three implementations now have a `Dockerfile` and need a host with a **persistent disk**, since each keeps its user accounts in a local SQLite file (`users.db`). That rules out pure static hosting and serverless functions (Vercel/Netlify functions, GitHub Pages) unless you swap in a hosted database. Good fits: **Render, Fly.io, Railway, or any VPS**.

General pattern for each folder:
```bash
cd <go|python|html>
docker build -t fifa-<folder> .
docker run -p <port>:<port> -v fifa_data:/data fifa-<folder>
# then seed a user against the same volume — see each folder's README for the exact command
```

See `go/README.md`, `python/README.md`, and `html/README.md` for folder-specific deploy commands and environment variables.

---

## 🔒 Security Notes

- **No demo/hardcoded credentials anywhere** — every implementation requires you to create a real user via its seed script before first login.
- **Passwords are bcrypt-hashed** in all three, stored in a local SQLite file that's git-ignored (never commit `users.db`).
- **Go:** real session cookies (HttpOnly, SameSite=Lax, Secure over HTTPS) with server-side middleware guarding `/dashboard` and every `/api/*` route except `/api/login`. Per-IP rate limiting on login (5/min).
- **Python:** Streamlit's `session_state` gate is enforced server-side (it can't be bypassed client-side like `localStorage` can); adds a per-session lockout after 5 failed attempts/minute.
- **Standalone (html):** previously had a fake client-side-only login (hardcoded credentials in the page's JS, "auth" stored in `localStorage`) that was trivially bypassable via view-source or devtools. That's been removed — the simulator page is now only served by the Express backend after a real session check.
- **Still worth doing before a public/production deploy:** HTTPS (via your host or a reverse proxy), CSRF protection on the state-changing endpoints, and a stronger rate-limit/lockout store shared across restarts (all three limiters are in-memory or per-browser-session today).

---

## 📚 For More Details

- **Go Implementation:** See `go/README.md`
- **Python Implementation:** See `python/README.md`
- **Standalone Implementation:** See `html/README.md`

---

## 👨‍💻 Created by
**Cikgu Faiz Academy**

---

## 📄 License
Educational project for learning purposes

---

## 🎯 Next Steps

1. **Pick your implementation** based on your needs:
   - Choose **Go** if you want speed and production-ready code
   - Choose **Python** if you want beautiful interactive UI
   - Choose **Standalone** if you want the simplest client-side simulator behind a minimal login server

2. **Read the README** in the folder of your choice

3. **Run it locally** and explore the code

4. **Customize it** - Change team Elo ratings, add new features, modify the UI!

Happy coding! ⚽🏆
