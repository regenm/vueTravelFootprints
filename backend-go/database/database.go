package database

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"travel-footprints/models"

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
	query := `
	CREATE TABLE IF NOT EXISTS markers (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		longitude TEXT NOT NULL,
		latitude TEXT NOT NULL,
		photos TEXT DEFAULT '[]',
		category TEXT DEFAULT '',
		notes TEXT DEFAULT '',
		visit_date TEXT DEFAULT '',
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	`
	_, err := db.conn.Exec(query)
	return err
}

func (db *DB) GetAllMarkers() ([]models.Marker, error) {
	rows, err := db.conn.Query(`SELECT id, name, longitude, latitude, photos, category, notes, visit_date, created_at, updated_at FROM markers ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMarkers(rows)
}

func (db *DB) GetMarkerByID(id string) (*models.Marker, error) {
	row := db.conn.QueryRow(`SELECT id, name, longitude, latitude, photos, category, notes, visit_date, created_at, updated_at FROM markers WHERE id = ?`, id)
	return scanMarker(row)
}

func (db *DB) CreateMarker(m models.Marker) error {
	photosJSON, _ := json.Marshal(m.Photos)
	_, err := db.conn.Exec(
		`INSERT INTO markers (id, name, longitude, latitude, photos, category, notes, visit_date, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.Name, m.Longitude, m.Latitude, string(photosJSON), m.Category, m.Notes, m.VisitDate, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

func (db *DB) UpdateMarker(id string, req models.UpdateMarkerRequest) (*models.Marker, error) {
	existing, err := db.GetMarkerByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Longitude != nil {
		existing.Longitude = *req.Longitude
	}
	if req.Latitude != nil {
		existing.Latitude = *req.Latitude
	}
	if req.Photos != nil {
		existing.Photos = req.Photos
	}
	if req.Category != nil {
		existing.Category = *req.Category
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	if req.VisitDate != nil {
		existing.VisitDate = *req.VisitDate
	}

	photosJSON, _ := json.Marshal(existing.Photos)
	_, err = db.conn.Exec(
		`UPDATE markers SET name=?, longitude=?, latitude=?, photos=?, category=?, notes=?, visit_date=?, updated_at=? WHERE id=?`,
		existing.Name, existing.Longitude, existing.Latitude, string(photosJSON), existing.Category, existing.Notes, existing.VisitDate, existing.UpdatedAt, id,
	)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

func (db *DB) DeleteMarker(id string) error {
	_, err := db.conn.Exec(`DELETE FROM markers WHERE id = ?`, id)
	return err
}

func (db *DB) SearchMarkers(keyword string, category string, startDate string, endDate string) ([]models.Marker, error) {
	query := `SELECT id, name, longitude, latitude, photos, category, notes, visit_date, created_at, updated_at FROM markers WHERE 1=1`
	args := []interface{}{}

	if keyword != "" {
		query += ` AND (name LIKE ? OR notes LIKE ?)`
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}
	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	if startDate != "" {
		query += ` AND visit_date >= ?`
		args = append(args, startDate)
	}
	if endDate != "" {
		query += ` AND visit_date <= ?`
		args = append(args, endDate)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMarkers(rows)
}

func scanMarker(row interface{ Scan(...interface{}) error }) (*models.Marker, error) {
	var m models.Marker
	var photosStr string

	err := row.Scan(&m.ID, &m.Name, &m.Longitude, &m.Latitude, &photosStr, &m.Category, &m.Notes, &m.VisitDate, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if photosStr != "" {
		json.Unmarshal([]byte(photosStr), &m.Photos)
	}
	if m.Photos == nil {
		m.Photos = []string{}
	}

	return &m, nil
}

func scanMarkers(rows *sql.Rows) ([]models.Marker, error) {
	markers := []models.Marker{}
	for rows.Next() {
		var m models.Marker
		var photosStr string
		err := rows.Scan(&m.ID, &m.Name, &m.Longitude, &m.Latitude, &photosStr, &m.Category, &m.Notes, &m.VisitDate, &m.CreatedAt, &m.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if photosStr != "" {
			json.Unmarshal([]byte(photosStr), &m.Photos)
		}
		if m.Photos == nil {
			m.Photos = []string{}
		}
		markers = append(markers, m)
	}
	return markers, nil
}

func (db *DB) InsertSeedData() error {
	var count int
	db.conn.QueryRow(`SELECT COUNT(*) FROM markers`).Scan(&count)
	if count > 0 {
		return nil
	}

	seeds := []struct {
		name     string
		lng      string
		lat      string
		photos   string
		category string
		notes    string
		date     string
	}{
		{"西湖", "120.15507", "30.27415", `["https://images.unsplash.com/photo-1506748686214-e9df14d4d9d0","https://images.unsplash.com/photo-1507525428034-b723cf961d3e"]`, "自然风光", "西湖真的是人间天堂，清晨的苏堤最美。", "2024-03-15"},
		{"黄山", "118.1689", "30.1336", `["https://images.unsplash.com/photo-1470770841072-f978cf4d019e","https://images.unsplash.com/photo-1501785888041-af3ef285b470"]`, "自然风光", "五岳归来不看山，黄山归来不看岳。", "2024-05-20"},
		{"长城", "116.5704", "40.4319", `["https://images.unsplash.com/photo-1501785888041-af3ef285b470","https://images.unsplash.com/photo-1491553895911-0055eca6402d"]`, "历史古迹", "不到长城非好汉！", "2024-07-10"},
	}

	for _, s := range seeds {
		m := models.NewMarker(models.CreateMarkerRequest{
			Name:      s.name,
			Longitude: s.lng,
			Latitude:  s.lat,
			Photos:    parsePhotoJSON(s.photos),
			Category:  s.category,
			Notes:     s.notes,
			VisitDate: s.date,
		})
		db.CreateMarker(m)
	}

	return nil
}

func parsePhotoJSON(s string) []string {
	var photos []string
	json.Unmarshal([]byte(s), &photos)
	if photos == nil {
		photos = []string{}
	}
	return photos
}