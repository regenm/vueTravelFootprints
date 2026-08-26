package database

import (
	"database/sql"
	"strings"
	"travel-footprints/models"
)

func (db *DB) CreateUser(u models.User) error {
	role := u.Role
	if role == "" {
		role = "user"
	}
	_, err := db.conn.Exec(
		`INSERT INTO users (id, username, email, password_hash, display_name, avatar, role, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		u.ID, strings.ToLower(u.Username), strings.ToLower(u.Email), u.PasswordHash, u.DisplayName, u.Avatar, role, u.CreatedAt, u.UpdatedAt,
	)
	return err
}

func (db *DB) GetUserByID(id string) (*models.User, error) {
	row := db.conn.QueryRow(
		`SELECT id, username, email, password_hash, display_name, avatar, role, created_at, updated_at FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	row := db.conn.QueryRow(
		`SELECT id, username, email, password_hash, display_name, avatar, role, created_at, updated_at FROM users WHERE username = ?`,
		strings.ToLower(username),
	)
	return scanUser(row)
}

func (db *DB) GetUserByAccount(account string) (*models.User, error) {
	account = strings.ToLower(strings.TrimSpace(account))
	row := db.conn.QueryRow(
		`SELECT id, username, email, password_hash, display_name, avatar, role, created_at, updated_at
		 FROM users WHERE username = ? OR email = ?`,
		account, account,
	)
	return scanUser(row)
}

func (db *DB) ListUsers() ([]models.User, error) {
	rows, err := db.conn.Query(
		`SELECT id, username, email, password_hash, display_name, avatar, role, created_at, updated_at
		 FROM users ORDER BY created_at ASC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.User{}
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		if u != nil {
			list = append(list, *u)
		}
	}
	return list, nil
}

func (db *DB) UpdateUser(u *models.User) error {
	u.UpdatedAt = models.NowISO()
	_, err := db.conn.Exec(
		`UPDATE users SET display_name=?, avatar=?, updated_at=? WHERE id=?`,
		u.DisplayName, u.Avatar, u.UpdatedAt, u.ID,
	)
	return err
}

func (db *DB) UpdateUserRole(id, role string) error {
	_, err := db.conn.Exec(`UPDATE users SET role=?, updated_at=? WHERE id=?`, role, models.NowISO(), id)
	return err
}

func (db *DB) UpdateUserPassword(id, passwordHash string) error {
	_, err := db.conn.Exec(`UPDATE users SET password_hash=?, updated_at=? WHERE id=?`, passwordHash, models.NowISO(), id)
	return err
}

func (db *DB) DeleteUserCascade(username string) error {
	user, err := db.GetUserByUsername(username)
	if err != nil || user == nil {
		return err
	}

	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM share_markers WHERE marker_id IN (SELECT id FROM markers WHERE user_id = ?)`, user.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM markers WHERE user_id = ?`, user.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM share_markers WHERE share_id IN (SELECT id FROM shares WHERE owner_id = ?)`, user.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM share_members WHERE share_id IN (SELECT id FROM shares WHERE owner_id = ?)`, user.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM shares WHERE owner_id = ?`, user.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM share_members WHERE user_id = ?`, user.ID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM users WHERE id = ?`, user.ID); err != nil {
		return err
	}
	return tx.Commit()
}

func scanUser(row interface{ Scan(...interface{}) error }) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.DisplayName, &u.Avatar, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if u.Role == "" {
		u.Role = "user"
	}
	return &u, nil
}
