# FIFA 2026 - Python Streamlit Backend

**Interactive tournament simulator built with Python Streamlit**

## Features
- 🎨 Beautiful interactive UI with Streamlit
- 🔐 Session-based authentication, backed by a SQLite user table (bcrypt-hashed passwords, no hardcoded credentials)
- 🚦 Per-session lockout after 5 failed login attempts within 60 seconds
- 📊 Real-time Monte Carlo simulations
- 📈 Interactive charts with Plotly
- 📝 Logging all login/logout events

## Project Structure
```
python/
├── app.py               # Streamlit app with login + simulator
├── auth_db.py            # SQLite schema + user creation/verification (bcrypt)
├── seed_user.py           # CLI to create/update a user (prompts for password)
├── login.html          # Standalone reference login markup (not served by app.py)
├── Dockerfile             # Container build for deployment
├── .env.example           # Config template
└── requirements.txt    # Python dependencies
```

## How to Run

### 1. Install Dependencies
```bash
pip install -r requirements.txt
```

### 2. Create a User

There are no default/demo credentials — create at least one user first:
```bash
cd python
python seed_user.py --username admin
# You'll be prompted for a password (min 8 characters, never echoed)
```
This stores a bcrypt hash in `users.db` (created automatically, git-ignored).

### 3. Run the Streamlit App
```bash
streamlit run app.py
```

### 4. Access the App
- **URL:** `http://localhost:8501`
- Your browser will automatically open the app

## Deployment

Build and run the Docker image on a host with a persistent disk (Render, Fly.io, Railway, a VPS):
```bash
docker build -t fifa-python .
docker run -p 8501:8501 -v fifa_data:/data fifa-python
```
Seed a user against the same volume before or after starting the container:
```bash
docker run --rm -v fifa_data:/data -e DB_PATH=/data/users.db -it fifa-python python seed_user.py --username admin
```
**Important:** without a persistent volume at `/data`, `users.db` (and every account) is wiped on every redeploy or restart.

## Features in the Dashboard
- **Login Page** - Session-based authentication
- **Welcome Message** - Shows logged-in username
- **Logout Button** - Quick logout with confirmation
- **Simulation Parameters**
  - Adjust Elo ratings for each team
  - Control number of simulations (1,000 - 100,000)
- **Results Visualization**
  - Championship probability bar chart
  - Top 4 teams KPI display
  - Favorite team highlight

## File Descriptions

### app.py
Contains:
- Login system using `st.session_state`
- Credential validation with demo users
- Elo-based win probability calculations
- Tournament simulation logic
- Interactive charts with Plotly
- Logging with Python's `logging` module

## Technologies
- **Framework:** Streamlit
- **Language:** Python 3.8+
- **Data Processing:** Pandas
- **Visualization:** Plotly
- **Logging:** Python logging module

## Dependencies
```
streamlit>=1.28.0
pandas>=2.0.0
plotly>=5.0.0
```

## Install & Run (One Command)
```bash
pip install streamlit pandas plotly && streamlit run app.py
```

## Performance
- Simulates 10,000 tournaments in < 1 second
- Responsive UI with real-time updates
- Efficient session state management
