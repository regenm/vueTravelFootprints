package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"travel-footprints/models"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1)
	_, _ = conn.Exec("PRAGMA foreign_keys = ON")
	_, _ = conn.Exec("PRAGMA busy_timeout = 5000")

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			display_name TEXT DEFAULT '',
			avatar TEXT DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS markers (
			id TEXT PRIMARY KEY,
			user_id TEXT DEFAULT '',
			name TEXT NOT NULL,
			longitude TEXT NOT NULL,
			latitude TEXT NOT NULL,
			photos TEXT DEFAULT '[]',
			category TEXT DEFAULT '',
			notes TEXT DEFAULT '',
			visit_date TEXT DEFAULT '',
			address TEXT DEFAULT '',
			is_public INTEGER DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS shares (
			id TEXT PRIMARY KEY,
			owner_id TEXT NOT NULL,
			token TEXT UNIQUE NOT NULL,
			title TEXT DEFAULT '',
			description TEXT DEFAULT '',
			type TEXT DEFAULT 'all',
			permission TEXT DEFAULT 'view',
			is_public INTEGER DEFAULT 1,
			expires_at TEXT DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS share_markers (
			share_id TEXT NOT NULL,
			marker_id TEXT NOT NULL,
			PRIMARY KEY (share_id, marker_id)
		)`,
		`CREATE TABLE IF NOT EXISTS share_members (
			share_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			permission TEXT DEFAULT 'view',
			created_at TEXT NOT NULL,
			PRIMARY KEY (share_id, user_id)
		)`,
	}

	for _, q := range stmts {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	if err := db.ensureColumn("markers", "user_id", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := db.ensureColumn("markers", "address", "TEXT DEFAULT ''"); err != nil {
		return err
	}
	if err := db.ensureColumn("markers", "is_public", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := db.ensureColumn("users", "role", "TEXT DEFAULT 'user'"); err != nil {
		return err
	}
	_, _ = db.conn.Exec(`UPDATE users SET role = 'user' WHERE role IS NULL OR role = ''`)

	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_markers_user ON markers(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_shares_token ON shares(token)`,
		`CREATE INDEX IF NOT EXISTS idx_shares_owner ON shares(owner_id)`,
		`CREATE INDEX IF NOT EXISTS idx_share_members_user ON share_members(user_id)`,
	}
	for _, q := range indexes {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("migrate index: %w", err)
		}
	}

	return nil
}

func (db *DB) ensureColumn(table, column, decl string) error {
	if db.hasColumn(table, column) {
		return nil
	}
	_, err := db.conn.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, decl))
	return err
}

func (db *DB) hasColumn(table, column string) bool {
	rows, err := db.conn.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false
		}
		if name == column {
			return true
		}
	}
	return false
}

func (db *DB) EnsureAdmin(username, password string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" {
		username = "admin"
	}

	existing, err := db.GetUserByUsername(username)
	if err != nil {
		return err
	}
	if existing != nil {
		if existing.Role != "admin" {
			if err := db.UpdateUserRole(existing.ID, "admin"); err != nil {
				return err
			}
		}
		if password != "" && password != "admin123" {
			if bcrypt.CompareHashAndPassword([]byte(existing.PasswordHash), []byte("admin123")) == nil {
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					return err
				}
				return db.UpdateUserPassword(existing.ID, string(hash))
			}
		}
		return nil
	}

	if password == "" {
		return fmt.Errorf("未找到管理员账号，请设置 ADMIN_PASSWORD")
	}
	if len(password) < 10 {
		return fmt.Errorf("管理员密码至少 10 位")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := models.NewUser(username, username+"@travel.local", string(hash), "管理员")
	user.Role = "admin"
	if err := db.CreateUser(user); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			found, err := db.GetUserByUsername(username)
			if err != nil || found == nil {
				return err
			}
			return db.UpdateUserRole(found.ID, "admin")
		}
		return err
	}
	return nil
}

func (db *DB) EnsureUser(username, password, displayName, email, role string) error {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" || password == "" {
		return nil
	}
	existing, err := db.GetUserByUsername(username)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}
	if displayName == "" {
		displayName = username
	}
	if email == "" {
		email = username + "@travel.local"
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user := models.NewUser(username, email, string(hash), displayName)
	if role == "admin" {
		user.Role = "admin"
	} else {
		user.Role = "user"
	}
	if err := db.CreateUser(user); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil
		}
		return err
	}
	return nil
}
