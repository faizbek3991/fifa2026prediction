const path = require('path');
const crypto = require('crypto');
const Database = require('better-sqlite3');
const bcrypt = require('bcryptjs');

const DB_PATH = process.env.DB_PATH || path.join(__dirname, 'users.db');
const SESSION_TTL_MS = 24 * 60 * 60 * 1000;

const db = new Database(DB_PATH);
db.pragma('journal_mode = WAL');

db.exec(`
  CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
  );

  CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    username TEXT NOT NULL,
    expires_at INTEGER NOT NULL
  );
`);

function upsertUser(username, password) {
  const hash = bcrypt.hashSync(password, 12);
  db.prepare(
    `INSERT INTO users (username, password_hash) VALUES (?, ?)
     ON CONFLICT(username) DO UPDATE SET password_hash = excluded.password_hash`
  ).run(username, hash);
}

// Compared against on every login attempt for a username that doesn't exist,
// so a lookup miss costs about as much as a real bcrypt comparison. Without
// this, unknown usernames would return almost instantly while real ones take
// the full bcrypt round-trip, letting an attacker enumerate valid usernames
// from response timing alone.
const DUMMY_HASH = bcrypt.hashSync('this-is-not-a-real-password-used-only-for-timing', 12);

function verifyCredentials(username, password) {
  const row = db.prepare('SELECT password_hash FROM users WHERE username = ?').get(username);
  const hashToCheck = row ? row.password_hash : DUMMY_HASH;
  const result = bcrypt.compareSync(password, hashToCheck);
  return result && Boolean(row);
}

function createSession(username) {
  const token = crypto.randomBytes(32).toString('hex');
  const expiresAt = Date.now() + SESSION_TTL_MS;
  db.prepare('INSERT INTO sessions (token, username, expires_at) VALUES (?, ?, ?)').run(
    token,
    username,
    expiresAt
  );
  return { token, expiresAt };
}

function getSessionUsername(token) {
  const row = db.prepare('SELECT username, expires_at FROM sessions WHERE token = ?').get(token);
  if (!row) return null;
  if (Date.now() > row.expires_at) {
    db.prepare('DELETE FROM sessions WHERE token = ?').run(token);
    return null;
  }
  return row.username;
}

function deleteSession(token) {
  db.prepare('DELETE FROM sessions WHERE token = ?').run(token);
}

module.exports = {
  SESSION_TTL_MS,
  upsertUser,
  verifyCredentials,
  createSession,
  getSessionUsername,
  deleteSession,
};
