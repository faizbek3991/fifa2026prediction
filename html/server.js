require('dotenv').config();
const path = require('path');
const express = require('express');
const cookieParser = require('cookie-parser');
const rateLimit = require('express-rate-limit');
const auth = require('./auth_db');

const app = express();
const PORT = process.env.PORT || 3000;
const SESSION_COOKIE = 'session_token';

app.use(express.json());
app.use(cookieParser());
app.disable('x-powered-by');

// Bootstrap a demo admin from env vars on every boot, so free-tier hosts
// without shell access (e.g. Render's free plan) don't need seed_user.js.
if (process.env.ADMIN_USERNAME && process.env.ADMIN_PASSWORD) {
  auth.upsertUser(process.env.ADMIN_USERNAME, process.env.ADMIN_PASSWORD);
}

function isSecureRequest(req) {
  return req.secure || req.headers['x-forwarded-proto'] === 'https';
}

function setSessionCookie(req, res, token, expiresAt) {
  res.cookie(SESSION_COOKIE, token, {
    httpOnly: true,
    secure: isSecureRequest(req),
    sameSite: 'lax',
    expires: new Date(expiresAt),
    path: '/',
  });
}

function clearSessionCookie(req, res) {
  res.clearCookie(SESSION_COOKIE, {
    httpOnly: true,
    secure: isSecureRequest(req),
    sameSite: 'lax',
    path: '/',
  });
}

function requireAuth(req, res, next) {
  const token = req.cookies[SESSION_COOKIE];
  const username = token && auth.getSessionUsername(token);
  if (!username) {
    if (req.path.startsWith('/api/')) {
      return res.status(401).json({ success: false, message: 'Unauthorized' });
    }
    return res.redirect('/');
  }
  req.username = username;
  next();
}

const loginLimiter = rateLimit({
  windowMs: 60 * 1000,
  max: 5,
  standardHeaders: true,
  legacyHeaders: false,
  message: { success: false, message: 'Too many attempts. Try again later.' },
});

app.get('/', (req, res) => {
  res.sendFile(path.join(__dirname, 'login.html'));
});

app.get('/dashboard', requireAuth, (req, res) => {
  res.sendFile(path.join(__dirname, 'wc2026_interactive_predictor.html'));
});

app.post('/api/login', loginLimiter, (req, res) => {
  const { username, password } = req.body || {};
  if (!username || !password) {
    return res.status(400).json({ success: false, message: 'Username and password required' });
  }

  const valid = auth.verifyCredentials(username, password);
  console.log(JSON.stringify({ time: new Date().toISOString(), event: 'login_attempt', username, success: valid, ip: req.ip }));

  if (!valid) {
    return res.status(401).json({ success: false, message: 'Invalid username or password' });
  }

  const { token, expiresAt } = auth.createSession(username);
  setSessionCookie(req, res, token, expiresAt);
  res.json({ success: true, message: 'Login successful' });
});

app.post('/api/register', loginLimiter, (req, res) => {
  const { username, password } = req.body || {};
  if (!username || !password) {
    return res.status(400).json({ success: false, message: 'Username and password required' });
  }
  if (username.length < 3 || username.length > 32) {
    return res.status(400).json({ success: false, message: 'Username must be 3-32 characters' });
  }
  if (password.length < 8) {
    return res.status(400).json({ success: false, message: 'Password must be at least 8 characters' });
  }

  const created = auth.createUser(username, password);
  console.log(JSON.stringify({ time: new Date().toISOString(), event: 'register_attempt', username, success: created, ip: req.ip }));

  if (!created) {
    return res.status(409).json({ success: false, message: 'Username already taken' });
  }

  res.json({ success: true, message: 'Account created. You can now log in.' });
});

app.post('/api/logout', requireAuth, (req, res) => {
  const token = req.cookies[SESSION_COOKIE];
  if (token) auth.deleteSession(token);
  clearSessionCookie(req, res);
  console.log(JSON.stringify({ time: new Date().toISOString(), event: 'logout', username: req.username, ip: req.ip }));
  res.json({ success: true, message: 'Logged out successfully' });
});

app.listen(PORT, () => {
  console.log(JSON.stringify({ time: new Date().toISOString(), event: 'server_started', port: PORT }));
});
