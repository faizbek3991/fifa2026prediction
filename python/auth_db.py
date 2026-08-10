"""SQLite-backed user store for the Streamlit login system."""
import os
import sqlite3
from contextlib import contextmanager

import bcrypt

DB_PATH = os.environ.get("DB_PATH", os.path.join(os.path.dirname(__file__), "users.db"))

SCHEMA = """
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
"""


@contextmanager
def _connect():
    conn = sqlite3.connect(DB_PATH)
    try:
        yield conn
    finally:
        conn.close()


def init_db():
    with _connect() as conn:
        conn.execute(SCHEMA)
        conn.commit()


def create_user(username: str, password: str) -> bool:
    """Create a user with a bcrypt-hashed password. Returns False if the username already exists."""
    password_hash = bcrypt.hashpw(password.encode("utf-8"), bcrypt.gensalt()).decode("utf-8")
    with _connect() as conn:
        try:
            conn.execute(
                "INSERT INTO users (username, password_hash) VALUES (?, ?)",
                (username, password_hash),
            )
            conn.commit()
            return True
        except sqlite3.IntegrityError:
            return False


def upsert_user(username: str, password: str) -> None:
    """Create the user, or update their password if the username already exists."""
    password_hash = bcrypt.hashpw(password.encode("utf-8"), bcrypt.gensalt()).decode("utf-8")
    with _connect() as conn:
        conn.execute(
            """
            INSERT INTO users (username, password_hash) VALUES (?, ?)
            ON CONFLICT(username) DO UPDATE SET password_hash = excluded.password_hash
            """,
            (username, password_hash),
        )
        conn.commit()


# Compared against on every login attempt for a username that doesn't exist,
# so a lookup miss costs about as much as a real bcrypt comparison. Without
# this, unknown usernames return almost instantly while real ones take the
# full bcrypt round-trip, letting an attacker enumerate valid usernames from
# response timing alone.
_DUMMY_HASH = bcrypt.hashpw(b"this-is-not-a-real-password-used-only-for-timing", bcrypt.gensalt())


def verify_credentials(username: str, password: str) -> bool:
    with _connect() as conn:
        row = conn.execute(
            "SELECT password_hash FROM users WHERE username = ?", (username,)
        ).fetchone()
    hash_to_check = row[0].encode("utf-8") if row is not None else _DUMMY_HASH
    result = bcrypt.checkpw(password.encode("utf-8"), hash_to_check)
    return result and row is not None


def bootstrap_admin_from_env() -> None:
    """Create/update a demo admin from env vars on every boot, so free-tier
    hosts without shell access (e.g. Render's free plan) don't need seed_user.py."""
    admin_user = os.environ.get("ADMIN_USERNAME")
    admin_pass = os.environ.get("ADMIN_PASSWORD")
    if admin_user and admin_pass:
        upsert_user(admin_user, admin_pass)


init_db()
bootstrap_admin_from_env()
