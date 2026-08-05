# WC2026 Interactive Predictor

**FIFA 2026 tournament prediction tool, now gated behind a real login**

## Overview
The Monte Carlo simulation itself is still pure client-side HTML/JS/Chart.js — it just no longer ships with a fake login screen. Client-side "auth" (a hardcoded credential list checked in the browser, gated by `localStorage`) can always be bypassed via view-source or devtools, so real gating now happens in a small Express server: `server.js` serves the login page, verifies credentials against a SQLite user table (bcrypt-hashed passwords), and only then serves `wc2026_interactive_predictor.html` from `/dashboard`.

## Files
- `server.js` — Express server: login/logout/session cookies + auth-gated `/dashboard`
- `auth_db.js` — SQLite schema + user/session helpers (bcrypt via `bcryptjs`)
- `seed_user.js` — CLI to create/update a user (prompts for password, hidden input)
- `login.html` — Login page served at `/`
- `wc2026_interactive_predictor.html` — The simulator itself, served only to authenticated sessions
- `Dockerfile`, `.env.example` — Deployment config

## How to Run

### 1. Install dependencies
```bash
cd html
npm install
```

### 2. Create a user
```bash
npm run seed-user -- --username admin
# prompts for a password (min 8 chars, hidden input)
```
This stores a bcrypt hash in `users.db` (created automatically, git-ignored).

### 3. Start the server
```bash
npm start
```

### 4. Access the app
- **Login:** `http://localhost:3000/`
- **Dashboard (requires login):** `http://localhost:3000/dashboard`

## Deployment

Needs a persistent-disk host (Render, Fly.io, Railway, a VPS) — not a pure static host or serverless functions, since it now has a real backend and a SQLite file that must survive restarts:
```bash
docker build -t fifa-html .
docker run -p 3000:3000 -v fifa_data:/data fifa-html
docker run --rm -v fifa_data:/data -e DB_PATH=/data/users.db -it fifa-html node seed_user.js --username admin
```

## Features
- 📊 **Monte Carlo Simulation** - Runs tournaments up to 200,000 iterations
- 🎚️ **Interactive Sliders** - Adjust team Elo ratings in real-time
- 📈 **Multiple Charts**
  - Championship probability (bar chart)
  - Stage progression (semi-final → final → champion)
  - Most likely final matchups
- 🏆 **H2H Banner** - Shows most likely final matchup with win percentages
- ⚡ **Performance Metrics** - Shows execution time in milliseconds

## Technologies
- **HTML5** - Structure
- **CSS3** - Styling with custom variables for easy theming
- **JavaScript** - Tournament simulation logic
- **Chart.js** - Data visualization library

## Customization
You can easily customize:
- Team Elo ratings (find `BASE` object in the script)
- Form adjustments (find `FORM` object)
- Bracket structure (modify `simulateTournament()` function)
- Color scheme (modify CSS variables in `:root`)

## Color Scheme Variables
```css
:root {
  --pitch: #07130c;        /* Dark background */
  --panel: #0d1f14;        /* Panel background */
  --line: #1c3a27;         /* Border color */
  --gold: #e9b949;         /* Accent color */
  --teal: #2fae8f;         /* Highlight color */
  --ink: #e8f0ea;          /* Text color */
  --dim: #8fa898;          /* Muted text */
}
```

## Browser Support
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+
