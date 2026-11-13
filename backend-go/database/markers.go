package database

import (
	"database/sql"
	"encoding/json"
	"travel-footprints/models"
)

const markerSelect = `SELECT id, user_id, name, longitude, latitude, photos, category, notes, visit_date, address, is_public, province_code, city_code, province_name, city_name, district_code, district_name, trip_id, created_at, updated_at FROM markers`

func (db *DB) GetMarkersByUser(userID string) ([]models.Marker, error) {
	rows, err := db.conn.Query(markerSelect+` WHERE user_id = ? ORDER BY COALESCE(NULLIF(visit_date,''), created_at) DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMarkers(rows)
}

func (db *DB) GetMarkerByID(id string) (*models.Marker, error) {
	row := db.conn.QueryRow(markerSelect+` WHERE id = ?`, id)
	return scanMarker(row)
}

func (db *DB) CreateMarker(m models.Marker) error {
	photosJSON, _ := json.Marshal(m.Photos)
	isPublic := 0
	if m.IsPublic {
		isPublic = 1
	}
	_, err := db.conn.Exec(
		`INSERT INTO markers (id, user_id, name, longitude, latitude, photos, category, notes, visit_date, address, is_public, province_code, city_code, province_name, city_name, district_code, district_name, trip_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.UserID, m.Name, m.Longitude, m.Latitude, string(photosJSON), m.Category, m.Notes, m.VisitDate, m.Address, isPublic, m.ProvinceCode, m.CityCode, m.ProvinceName, m.CityName, m.DistrictCode, m.DistrictName, m.TripID, m.CreatedAt, m.UpdatedAt,
	)
	return err
}

func (db *DB) UpdateMarker(id string, req models.UpdateMarkerRequest) (*models.Marker, error) {
	existing, err := db.GetMarkerByID(id)
	if err != nil || existing == nil {
		return existing, err
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
	if req.Address != nil {
		existing.Address = *req.Address
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
	if req.TripID != nil {
		existing.TripID = *req.TripID
	}
	existing.UpdatedAt = models.NowISO()

	photosJSON, _ := json.Marshal(existing.Photos)
	isPublic := 0
	if existing.IsPublic {
		isPublic = 1
	}
	_, err = db.conn.Exec(
		`UPDATE markers SET name=?, longitude=?, latitude=?, photos=?, category=?, notes=?, visit_date=?, address=?, is_public=?, province_code=?, city_code=?, province_name=?, city_name=?, district_code=?, district_name=?, trip_id=?, updated_at=? WHERE id=?`,
		existing.Name, existing.Longitude, existing.Latitude, string(photosJSON), existing.Category, existing.Notes, existing.VisitDate, existing.Address, isPublic, existing.ProvinceCode, existing.CityCode, existing.ProvinceName, existing.CityName, existing.DistrictCode, existing.DistrictName, existing.TripID, existing.UpdatedAt, id,
	)
	if err != nil {
		return nil, err
	}
	return existing, nil
}

func (db *DB) DeleteMarker(id string) error {
	_, _ = db.conn.Exec(`DELETE FROM share_markers WHERE marker_id = ?`, id)
	_, err := db.conn.Exec(`DELETE FROM markers WHERE id = ?`, id)
	return err
}

// CountMarkersReferencingPhoto 统计引用该图片 URL 的其他足迹数量（photos 以 JSON 数组存储，按带引号精确匹配）
func (db *DB) CountMarkersReferencingPhoto(photoURL, excludeMarkerID string) (int, error) {
	var n int
	err := db.conn.QueryRow(
		`SELECT COUNT(*) FROM markers WHERE photos LIKE ? AND id != ?`,
		`%"`+photoURL+`"%`, excludeMarkerID,
	).Scan(&n)
	return n, err
}

func (db *DB) SearchMarkers(userID, keyword, category, startDate, endDate string) ([]models.Marker, error) {
	query := markerSelect + ` WHERE user_id = ?`
	args := []interface{}{userID}

	if keyword != "" {
		query += ` AND (name LIKE ? OR notes LIKE ? OR address LIKE ?)`
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
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
	query += ` ORDER BY COALESCE(NULLIF(visit_date,''), created_at) DESC`

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
	var isPublic int
	err := row.Scan(&m.ID, &m.UserID, &m.Name, &m.Longitude, &m.Latitude, &photosStr, &m.Category, &m.Notes, &m.VisitDate, &m.Address, &isPublic, &m.ProvinceCode, &m.CityCode, &m.ProvinceName, &m.CityName, &m.DistrictCode, &m.DistrictName, &m.TripID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	m.IsPublic = isPublic == 1
	m.Photos = decodePhotos(photosStr)
	return &m, nil
}

func scanMarkers(rows *sql.Rows) ([]models.Marker, error) {
	markers := []models.Marker{}
	for rows.Next() {
		var m models.Marker
		var photosStr string
		var isPublic int
		if err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.Longitude, &m.Latitude, &photosStr, &m.Category, &m.Notes, &m.VisitDate, &m.Address, &isPublic, &m.ProvinceCode, &m.CityCode, &m.ProvinceName, &m.CityName, &m.DistrictCode, &m.DistrictName, &m.TripID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		m.IsPublic = isPublic == 1
		m.Photos = decodePhotos(photosStr)
		markers = append(markers, m)
	}
	return markers, nil
}

func decodePhotos(s string) []string {
	photos := []string{}
	if s != "" {
		_ = json.Unmarshal([]byte(s), &photos)
	}
	if photos == nil {
		photos = []string{}
	}
	return photos
}

func (db *DB) UpdateMarkerRegion(id string, region models.RegionInfo) error {
	_, err := db.conn.Exec(
		`UPDATE markers SET province_code=?, city_code=?, province_name=?, city_name=?, district_code=?, district_name=?, updated_at=? WHERE id=?`,
		region.ProvinceCode, region.CityCode, region.ProvinceName, region.CityName, region.DistrictCode, region.DistrictName, models.NowISO(), id,
	)
	return err
}
