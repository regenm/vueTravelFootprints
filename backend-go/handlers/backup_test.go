package handlers

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"travel-footprints/database"
	"travel-footprints/middleware"

	_ "modernc.org/sqlite"
)

func newBackupEnv(t *testing.T) (*BackupHandler, string, string) {
	t.Helper()
	dir := t.TempDir()
	db, err := database.New(filepath.Join(dir, "travel.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.EnsureAdmin("admin", "password123456"); err != nil {
		t.Fatal(err)
	}
	if err := db.EnsureUser("u1", "password123456", "普通用户", "u1@travel.local", "user"); err != nil {
		t.Fatal(err)
	}
	uploads := filepath.Join(dir, "uploads")
	if err := os.MkdirAll(uploads, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(uploads, "a.jpg"), []byte("fake-jpeg-bytes"), 0644); err != nil {
		t.Fatal(err)
	}
	admin, _ := db.GetUserByUsername("admin")
	user, _ := db.GetUserByUsername("u1")
	return NewBackupHandler(db, uploads), admin.ID, user.ID
}

func exportReq(uid string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/api/admin/export", nil)
	return r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, uid))
}

func readZip(t *testing.T, body *bytes.Buffer) map[string][]byte {
	t.Helper()
	zr, err := zip.NewReader(bytes.NewReader(body.Bytes()), int64(body.Len()))
	if err != nil {
		t.Fatalf("响应不是合法 zip: %v", err)
	}
	files := map[string][]byte{}
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, _ := io.ReadAll(rc)
		rc.Close()
		files[f.Name] = data
	}
	return files
}

func TestBackupExportAdmin(t *testing.T) {
	h, adminID, _ := newBackupEnv(t)
	rec := httptest.NewRecorder()
	h.Export(rec, exportReq(adminID))

	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 %d, body: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/zip" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.HasPrefix(cd, `attachment; filename="travel-backup-`) {
		t.Fatalf("Content-Disposition = %q", cd)
	}

	files := readZip(t, rec.Body)
	snap, ok := files["data/travel.db"]
	if !ok {
		t.Fatalf("zip 缺少 data/travel.db, 含: %v", keysOf(files))
	}
	extract := filepath.Join(t.TempDir(), "snap.db")
	if err := os.WriteFile(extract, snap, 0644); err != nil {
		t.Fatal(err)
	}
	conn, err := sql.Open("sqlite", extract)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var n int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM users WHERE username = 'admin'`).Scan(&n); err != nil {
		t.Fatalf("zip 内数据库不可用: %v", err)
	}
	if n != 1 {
		t.Fatalf("zip 内数据库缺少 admin, got %d", n)
	}

	img, ok := files["uploads/a.jpg"]
	if !ok || string(img) != "fake-jpeg-bytes" {
		t.Fatalf("zip 缺少或图片内容不符: uploads/a.jpg")
	}
}

func TestBackupExportForbidden(t *testing.T) {
	h, _, userID := newBackupEnv(t)
	rec := httptest.NewRecorder()
	h.Export(rec, exportReq(userID))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("普通用户应被拒绝, got %d", rec.Code)
	}

	rec2 := httptest.NewRecorder()
	h.Export(rec2, exportReq("no-such-user"))
	if rec2.Code != http.StatusForbidden {
		t.Fatalf("不存在用户应被拒绝, got %d", rec2.Code)
	}
}

func keysOf(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
