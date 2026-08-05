package main

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")

func openDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		expires_at DATETIME NOT NULL
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}
	return db, nil
}

func createUser(db *sql.DB, username, passwordHash string) error {
	_, err := db.Exec(`INSERT INTO users (username, password_hash) VALUES (?, ?)`, username, passwordHash)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return ErrUserExists
		}
		return err
	}
	return nil
}

func getPasswordHash(db *sql.DB, username string) (string, error) {
	var hash string
	err := db.QueryRow(`SELECT password_hash FROM users WHERE username = ?`, username).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}
	return hash, nil
}

func createSession(db *sql.DB, token, username string, ttl time.Duration) error {
	_, err := db.Exec(`INSERT INTO sessions (token, username, expires_at) VALUES (?, ?, ?)`,
		token, username, time.Now().Add(ttl))
	return err
}

func getSessionUsername(db *sql.DB, token string) (string, error) {
	var username string
	var expiresAt time.Time
	err := db.QueryRow(`SELECT username, expires_at FROM sessions WHERE token = ?`, token).Scan(&username, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}
	if time.Now().After(expiresAt) {
		_, _ = db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
		return "", ErrUserNotFound
	}
	return username, nil
}

func deleteSession(db *sql.DB, token string) error {
	_, err := db.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	return err
}

func isUniqueConstraintErr(err error) bool {
	return err != nil && strings.Contains(err.Error(), "constraint failed")
}
