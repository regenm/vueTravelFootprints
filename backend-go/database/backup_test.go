package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSnapshotTo(t *testing.T) {
	dir := t.TempDir()
	db, err := New(filepath.Join(dir, "src.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.EnsureAdmin("admin", "password123456"); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(dir, "snapshot.db")
	if err := db.SnapshotTo(target); err != nil {
		t.Fatalf("SnapshotTo 失败: %v", err)
	}

	snap, err := sql.Open("sqlite", target)
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Close()
	var n int
	if err := snap.QueryRow(`SELECT COUNT(*) FROM users WHERE username = 'admin'`).Scan(&n); err != nil {
		t.Fatalf("快照不是可用的 SQLite 库: %v", err)
	}
	if n != 1 {
		t.Fatalf("快照缺少 admin 用户, got %d", n)
	}
}
