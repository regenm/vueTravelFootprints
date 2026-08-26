package database

import (
	"database/sql"
	"travel-footprints/models"
)

func (db *DB) CreateShare(s models.Share, markerIDs []string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	isPublic := 0
	if s.IsPublic {
		isPublic = 1
	}
	_, err = tx.Exec(
		`INSERT INTO shares (id, owner_id, token, title, description, type, permission, is_public, expires_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		s.ID, s.OwnerID, s.Token, s.Title, s.Description, s.Type, s.Permission, isPublic, s.ExpiresAt, s.CreatedAt,
	)
	if err != nil {
		return err
	}

	if s.Type == "selected" {
		for _, mid := range markerIDs {
			if _, err := tx.Exec(`INSERT OR IGNORE INTO share_markers (share_id, marker_id) VALUES (?, ?)`, s.ID, mid); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (db *DB) GetShareByID(id string) (*models.Share, error) {
	row := db.conn.QueryRow(
		`SELECT id, owner_id, token, title, description, type, permission, is_public, expires_at, created_at FROM shares WHERE id = ?`,
		id,
	)
	return scanShare(row)
}

func (db *DB) GetShareByToken(token string) (*models.Share, error) {
	row := db.conn.QueryRow(
		`SELECT id, owner_id, token, title, description, type, permission, is_public, expires_at, created_at FROM shares WHERE token = ?`,
		token,
	)
	return scanShare(row)
}

func (db *DB) ListSharesByOwner(ownerID string) ([]models.Share, error) {
	rows, err := db.conn.Query(
		`SELECT id, owner_id, token, title, description, type, permission, is_public, expires_at, created_at
		 FROM shares WHERE owner_id = ? ORDER BY created_at DESC`,
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanShares(rows)
}

func (db *DB) ListSharesForMember(userID string) ([]models.Share, error) {
	rows, err := db.conn.Query(
		`SELECT s.id, s.owner_id, s.token, s.title, s.description, s.type, s.permission, s.is_public, s.expires_at, s.created_at
		 FROM shares s
		 INNER JOIN share_members m ON m.share_id = s.id
		 WHERE m.user_id = ?
		 ORDER BY s.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanShares(rows)
}

func (db *DB) DeleteShare(id, ownerID string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`DELETE FROM shares WHERE id = ? AND owner_id = ?`, id, ownerID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	_, _ = tx.Exec(`DELETE FROM share_markers WHERE share_id = ?`, id)
	_, _ = tx.Exec(`DELETE FROM share_members WHERE share_id = ?`, id)
	return tx.Commit()
}

func (db *DB) AddShareMarker(shareID, markerID string) error {
	_, err := db.conn.Exec(`INSERT OR IGNORE INTO share_markers (share_id, marker_id) VALUES (?, ?)`, shareID, markerID)
	return err
}

func (db *DB) AddShareMember(shareID, userID, permission string) error {
	if permission != "edit" {
		permission = "view"
	}
	_, err := db.conn.Exec(
		`INSERT INTO share_members (share_id, user_id, permission, created_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(share_id, user_id) DO UPDATE SET permission=excluded.permission`,
		shareID, userID, permission, models.NowISO(),
	)
	return err
}

func (db *DB) RemoveShareMember(shareID, userID string) error {
	_, err := db.conn.Exec(`DELETE FROM share_members WHERE share_id = ? AND user_id = ?`, shareID, userID)
	return err
}

func (db *DB) UpdateShare(s *models.Share) error {
	isPublic := 0
	if s.IsPublic {
		isPublic = 1
	}
	_, err := db.conn.Exec(
		`UPDATE shares SET title=?, description=?, permission=?, is_public=? WHERE id=?`,
		s.Title, s.Description, s.Permission, isPublic, s.ID,
	)
	return err
}

func (db *DB) GetShareMembers(shareID string) ([]models.ShareMember, error) {
	rows, err := db.conn.Query(
		`SELECT m.user_id, u.username, u.display_name, u.avatar, m.permission, m.created_at
		 FROM share_members m
		 INNER JOIN users u ON u.id = m.user_id
		 WHERE m.share_id = ?
		 ORDER BY m.created_at ASC`,
		shareID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []models.ShareMember{}
	for rows.Next() {
		var mem models.ShareMember
		if err := rows.Scan(&mem.UserID, &mem.Username, &mem.DisplayName, &mem.Avatar, &mem.Permission, &mem.CreatedAt); err != nil {
			return nil, err
		}
		if mem.DisplayName == "" {
			mem.DisplayName = mem.Username
		}
		members = append(members, mem)
	}
	return members, nil
}

func (db *DB) GetMemberPermission(shareID, userID string) (string, error) {
	var perm string
	err := db.conn.QueryRow(
		`SELECT permission FROM share_members WHERE share_id = ? AND user_id = ?`,
		shareID, userID,
	).Scan(&perm)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return perm, err
}

func (db *DB) GetShareMarkers(s *models.Share) ([]models.Marker, error) {
	if s.Type == "all" {
		owned, err := db.GetMarkersByUser(s.OwnerID)
		if err != nil {
			return nil, err
		}
		extra, err := db.getMarkersByShareID(s.ID)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		merged := make([]models.Marker, 0, len(owned)+len(extra))
		for _, m := range owned {
			seen[m.ID] = true
			merged = append(merged, m)
		}
		for _, m := range extra {
			if !seen[m.ID] {
				merged = append(merged, m)
			}
		}
		return merged, nil
	}
	return db.getMarkersByShareID(s.ID)
}

func (db *DB) CountShareMarkers(s *models.Share) (int, error) {
	markers, err := db.GetShareMarkers(s)
	if err != nil {
		return 0, err
	}
	return len(markers), nil
}

func (db *DB) getMarkersByShareID(shareID string) ([]models.Marker, error) {
	rows, err := db.conn.Query(
		markerSelect+` WHERE id IN (SELECT marker_id FROM share_markers WHERE share_id = ?)
		 ORDER BY COALESCE(NULLIF(visit_date,''), created_at) DESC`,
		shareID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMarkers(rows)
}

func (db *DB) AttachAuthors(markers []models.Marker) []models.Marker {
	if len(markers) == 0 {
		return markers
	}
	cache := map[string]*models.UserPublic{}
	for i := range markers {
		uid := markers[i].UserID
		if uid == "" {
			continue
		}
		if pub, ok := cache[uid]; ok {
			markers[i].Author = pub
			continue
		}
		user, err := db.GetUserByID(uid)
		if err != nil || user == nil {
			continue
		}
		p := user.Public()
		cache[uid] = &p
		markers[i].Author = &p
	}
	return markers
}

func scanShare(row interface{ Scan(...interface{}) error }) (*models.Share, error) {
	var s models.Share
	var isPublic int
	err := row.Scan(&s.ID, &s.OwnerID, &s.Token, &s.Title, &s.Description, &s.Type, &s.Permission, &isPublic, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	s.IsPublic = isPublic == 1
	return &s, nil
}

func scanShares(rows *sql.Rows) ([]models.Share, error) {
	list := []models.Share{}
	for rows.Next() {
		var s models.Share
		var isPublic int
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.Token, &s.Title, &s.Description, &s.Type, &s.Permission, &isPublic, &s.ExpiresAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.IsPublic = isPublic == 1
		list = append(list, s)
	}
	return list, nil
}
