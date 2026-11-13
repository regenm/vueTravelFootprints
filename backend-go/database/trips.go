package database

import (
	"database/sql"
	"errors"
	"travel-footprints/models"
)

var errNoTrip = errors.New("行程不存在")

const tripSelect = `SELECT id, user_id, name, start_date, end_date, notes, created_at FROM trips`

func scanTrip(row interface{ Scan(...interface{}) error }) (*models.Trip, error) {
	var t models.Trip
	if err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.StartDate, &t.EndDate, &t.Notes, &t.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // 与 scanMarker 一致：未找到时返回 nil 而非错误，让调用方走 404 分支
		}
		return nil, err
	}
	return &t, nil
}

func (db *DB) GetTripsByUser(userID string) ([]models.Trip, error) {
	rows, err := db.conn.Query(tripSelect+` WHERE user_id = ? ORDER BY COALESCE(NULLIF(start_date,''), created_at) DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []models.Trip{}
	for rows.Next() {
		t, err := scanTrip(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *t)
	}
	return list, nil
}

func (db *DB) CreateTrip(t models.Trip) error {
	_, err := db.conn.Exec(
		`INSERT INTO trips (id, user_id, name, start_date, end_date, notes, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.UserID, t.Name, t.StartDate, t.EndDate, t.Notes, t.CreatedAt,
	)
	return err
}

func (db *DB) UpdateTrip(id, userID string, req models.UpdateTripRequest) (*models.Trip, error) {
	row := db.conn.QueryRow(tripSelect+` WHERE id = ? AND user_id = ?`, id, userID)
	existing, err := scanTrip(row)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.StartDate != nil {
		existing.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		existing.EndDate = *req.EndDate
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	_, err = db.conn.Exec(
		`UPDATE trips SET name=?, start_date=?, end_date=?, notes=? WHERE id=?`,
		existing.Name, existing.StartDate, existing.EndDate, existing.Notes, id,
	)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (db *DB) DeleteTrip(id, userID string) error {
	res, err := db.conn.Exec(`DELETE FROM trips WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return errNoTrip
	}
	_, _ = db.conn.Exec(`UPDATE markers SET trip_id = '' WHERE trip_id = ? AND user_id = ?`, id, userID)
	return nil
}
