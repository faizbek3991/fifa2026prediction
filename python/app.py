import streamlit as st
import pandas as pd
import plotly.express as px
import random
import logging
import time

import auth_db

# --- Logging Configuration ---
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# --- Page Configuration ---
st.set_page_config(page_title="FIFA 2026 Monte Carlo Predictor", layout="wide")

# --- Login System ---
MAX_ATTEMPTS = 5
LOCKOUT_WINDOW_SECONDS = 60

def init_session_state():
    """Initialize session state variables"""
    if 'logged_in' not in st.session_state:
        st.session_state.logged_in = False
    if 'username' not in st.session_state:
        st.session_state.username = None
    if 'login_attempts' not in st.session_state:
        st.session_state.login_attempts = []

def validate_credentials(username, password):
    """Validate login credentials against the SQLite user store (bcrypt-hashed passwords)."""
    is_valid = auth_db.verify_credentials(username, password)

    # Log login attempt
    logger.info(f"Login attempt - Username: {username}, Success: {is_valid}")

    return is_valid

def is_rate_limited():
    """Per-browser-session throttle: MAX_ATTEMPTS failed logins per LOCKOUT_WINDOW_SECONDS."""
    now = time.time()
    st.session_state.login_attempts = [
        t for t in st.session_state.login_attempts if now - t < LOCKOUT_WINDOW_SECONDS
    ]
    return len(st.session_state.login_attempts) >= MAX_ATTEMPTS

def logout():
    """Handle logout"""
    logger.info(f"User logged out - Username: {st.session_state.username}")
    st.session_state.logged_in = False
    st.session_state.username = None

def login_page():
    """Display login page"""
    _, col_center, _ = st.columns([1, 2, 1])

    with col_center:
        st.markdown("<h1 style='text-align: center; color: #4CAF50;'>⚽ FIFA 2026 Simulator</h1>", unsafe_allow_html=True)
        st.markdown("<p style='text-align: center; color: #666;'>Tournament Prediction Engine</p>", unsafe_allow_html=True)
        st.markdown("---")

        username = st.text_input("Username", placeholder="Enter your username")
        password = st.text_input("Password", type="password", placeholder="Enter your password")

        if st.button("Login", use_container_width=True, type="primary"):
            if is_rate_limited():
                st.error("Too many failed attempts. Please wait a minute and try again.")
            elif validate_credentials(username, password):
                st.session_state.logged_in = True
                st.session_state.username = username
                st.session_state.login_attempts = []
                logger.info(f"User logged in successfully - Username: {username}")
                st.success("Login successful! Redirecting...")
                st.rerun()
            else:
                st.session_state.login_attempts.append(time.time())
                logger.warning(f"Failed login attempt - Username: {username}")
                st.error("Invalid username or password")

# Initialize session state
init_session_state()

# Check if user is logged in
if not st.session_state.logged_in:
    login_page()
    st.stop()

# --- Dashboard Navbar ---
col1, col2, col3 = st.columns([0.7, 0.2, 0.1])
with col1:
    st.markdown(f"### ⚽ FIFA 2026 Simulator - Welcome, {st.session_state.username}!")
with col3:
    if st.button("🚪 Logout", use_container_width=True):
        logout()
        st.rerun()

st.markdown("---")

# --- Sidebar for User Controls ---
st.sidebar.header("⚙️ Simulation Parameters")
st.sidebar.markdown("Adjust the Elo ratings or the number of simulations.")

# Default Elo Ratings (as of July 2026)
default_elos = {
    'France': 2143, 'Spain': 2177, 'Belgium': 1961,
    'Norway': 1972, 'England': 2076, 'Argentina': 2156, 'Switzerland': 1943
}

# Create sliders for Elo ratings
elo_ratings = {}
st.sidebar.subheader("Team Elo Ratings")
for team, default_rating in default_elos.items():
    elo_ratings[team] = st.sidebar.slider(team, 1500, 2500, default_rating, 10)

# Simulation count slider
num_simulations = st.sidebar.slider("Number of Simulations", 1000, 100000, 10000, 1000)

# --- Core Simulation Logic ---
def get_win_probability(rating_a, rating_b):
    return 1 / (1 + 10 ** ((rating_b - rating_a) / 400))

def simulate_match(team_a, team_b):
    prob_a = get_win_probability(elo_ratings[team_a], elo_ratings[team_b])
    return team_a if random.random() < prob_a else team_b

def simulate_tournament():
    # Quarter-finals
    qf2_winner = simulate_match('Spain', 'Belgium')
    qf3_winner = simulate_match('Norway', 'England')
    qf4_winner = simulate_match('Argentina', 'Switzerland')
    
    # Semi-finals
    sf1_winner = simulate_match('France', qf2_winner)
    sf2_winner = simulate_match(qf3_winner, qf4_winner)
    
    # Final (Upgraded to return the Champion)
    champion = simulate_match(sf1_winner, sf2_winner)
    return champion

# --- Main Dashboard UI ---
st.title("🏆 FIFA 2026 World Cup Monte Carlo Predictor")
st.markdown("Simulating the remaining knockout stages to predict the ultimate champion.")

# Run button
if st.button("🚀 Run Simulation", type="primary", use_container_width=True):
    with st.spinner(f"Running {num_simulations:,} simulations..."):
        champions_count = {team: 0 for team in elo_ratings}
        
        for _ in range(num_simulations):
            winner = simulate_tournament()
            champions_count[winner] += 1
            
        # Convert to DataFrame for plotting
        df = pd.DataFrame({
            'Team': list(champions_count.keys()),
            'Probability (%)': [ (count / num_simulations) * 100 for count in champions_count.values() ]
        })
        df = df.sort_values(by='Probability (%)', ascending=False).reset_index(drop=True)
        
    # Display Results
    st.success("Simulation Complete!")
    
    col1, col2 = st.columns([2, 1])
    
    with col1:
        st.subheader("Probability of Winning the Tournament")
        fig = px.bar(
            df, 
            x='Team', 
            y='Probability (%)', 
            color='Probability (%)',
            color_continuous_scale='Viridis',
            text_auto='.2f'
        )
        fig.update_layout(xaxis_title=None, yaxis_title="Win Probability (%)")
        st.plotly_chart(fig, use_container_width=True)
        
    with col2:
        st.subheader("Raw Data")
        st.dataframe(df, use_container_width=True, hide_index=True)
        
        # Highlight the favorite
        favorite = df.iloc[0]['Team']
        fav_prob = df.iloc[0]['Probability (%)']
        st.info(f"🌟 **Favorite to win:** {favorite} ({fav_prob:.2f}%)")

else:
    st.info("Adjust the parameters in the sidebar and click **Run Simulation** to start.")