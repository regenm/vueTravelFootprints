package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"travel-footprints/database"
	"travel-footprints/middleware"
	"travel-footprints/models"
)

func newMarkerEnv(t *testing.T) (*MarkerHandler, *database.DB, string, string) {
	t.Helper()
	dir := t.TempDir()
	db, err := database.New(filepath.Join(dir, "travel.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.EnsureUser("u1", "password123456", "用户一", "u1@travel.local", "user"); err != nil {
		t.Fatal(err)
	}
	uploads := filepath.Join(dir, "uploads")
	if err := os.MkdirAll(uploads, 0755); err != nil {
		t.Fatal(err)
	}
	user, _ := db.GetUserByUsername("u1")
	return NewMarkerHandler(db, NewGeocodeClient(""), uploads), db, user.ID, uploads
}

func writePhoto(t *testing.T, uploads, name string) string {
	t.Helper()
	if err := os.WriteFile(filepath.Join(uploads, name), []byte("img-"+name), 0644); err != nil {
		t.Fatal(err)
	}
	return "/uploads/" + name
}

func createMarkerWithPhotos(t *testing.T, db *database.DB, uid string, photos []string) *models.Marker {
	t.Helper()
	m := models.NewMarker(uid, models.CreateMarkerRequest{
		Name: "测试足迹", Longitude: "121.0", Latitude: "28.0", Photos: photos,
	})
	if err := db.CreateMarker(m); err != nil {
		t.Fatal(err)
	}
	// 预填行政区划，避免 Update 时触发外部地理编码请求
	if err := db.UpdateMarkerRegion(m.ID, models.RegionInfo{ProvinceCode: "330000", CityCode: "331000", ProvinceName: "浙江省", CityName: "台州市"}); err != nil {
		t.Fatal(err)
	}
	return &m
}

func serveMarkerReq(h *MarkerHandler, method, path, body string, uid string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/markers/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/markers/{id}", h.Update)
	mux.HandleFunc("DELETE /api/markers/{id}", h.Delete)
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, uid))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func photoExists(uploads, name string) bool {
	_, err := os.Stat(filepath.Join(uploads, name))
	return err == nil
}

func TestDeleteMarkerCleansUnreferencedPhotos(t *testing.T) {
	h, db, uid, uploads := newMarkerEnv(t)
	a := writePhoto(t, uploads, "u1_100.jpg")
	b := writePhoto(t, uploads, "u1_200.jpg")
	m := createMarkerWithPhotos(t, db, uid, []string{a, b})

	rec := serveMarkerReq(h, http.MethodDelete, "/api/markers/"+m.ID, "", uid)
	if rec.Code != http.StatusOK {
		t.Fatalf("删除足迹失败: %d %s", rec.Code, rec.Body.String())
	}
	for _, name := range []string{"u1_100.jpg", "u1_200.jpg"} {
		if photoExists(uploads, name) {
			t.Fatalf("删除足迹后图片应被清理: %s", name)
		}
	}
}

func TestDeleteMarkerKeepsPhotosReferencedByOtherMarker(t *testing.T) {
	h, db, uid, uploads := newMarkerEnv(t)
	shared := writePhoto(t, uploads, "u1_shared.jpg")
	m1 := createMarkerWithPhotos(t, db, uid, []string{shared})
	createMarkerWithPhotos(t, db, uid, []string{shared})

	rec := serveMarkerReq(h, http.MethodDelete, "/api/markers/"+m1.ID, "", uid)
	if rec.Code != http.StatusOK {
		t.Fatalf("删除足迹失败: %d %s", rec.Code, rec.Body.String())
	}
	if !photoExists(uploads, "u1_shared.jpg") {
		t.Fatal("图片仍被其他足迹引用，不应删除")
	}
}

func TestUpdateMarkerCleansRemovedPhotos(t *testing.T) {
	h, db, uid, uploads := newMarkerEnv(t)
	a := writePhoto(t, uploads, "u1_a.jpg")
	b := writePhoto(t, uploads, "u1_b.jpg")
	m := createMarkerWithPhotos(t, db, uid, []string{a, b})

	body := `{"photos":["/uploads/u1_a.jpg"]}`
	rec := serveMarkerReq(h, http.MethodPut, "/api/markers/"+m.ID, body, uid)
	if rec.Code != http.StatusOK {
		t.Fatalf("更新足迹失败: %d %s", rec.Code, rec.Body.String())
	}
	if !photoExists(uploads, "u1_a.jpg") {
		t.Fatal("仍被引用的图片不应删除")
	}
	if photoExists(uploads, "u1_b.jpg") {
		t.Fatal("编辑中移除的图片应被清理")
	}
}
