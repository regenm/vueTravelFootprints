package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"travel-footprints/database"
	"travel-footprints/middleware"
)

func newAuthEnv(t *testing.T) (*AuthHandler, string) {
	t.Helper()
	db, err := database.New(filepath.Join(t.TempDir(), "travel.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.EnsureAdmin("admin", "password123456"); err != nil {
		t.Fatal(err)
	}
	admin, _ := db.GetUserByUsername("admin")
	return NewAuthHandler(db, "test-secret"), admin.ID
}

func changePasswordReq(h *AuthHandler, uid, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(http.MethodPut, "/api/auth/password", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDKey, uid))
	rec := httptest.NewRecorder()
	h.ChangePassword(rec, r)
	return rec
}

func loginReq(h *AuthHandler, account, password string) int {
	body := `{"account":"` + account + `","password":"` + password + `"}`
	r := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Login(rec, r)
	return rec.Code
}

func TestChangePasswordOK(t *testing.T) {
	h, uid := newAuthEnv(t)

	rec := changePasswordReq(h, uid, `{"oldPassword":"password123456","newPassword":"newpassword99"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("修改密码失败: %d %s", rec.Code, rec.Body.String())
	}
	if code := loginReq(h, "admin", "newpassword99"); code != http.StatusOK {
		t.Fatalf("新密码应能登录, got %d", code)
	}
	if code := loginReq(h, "admin", "password123456"); code != http.StatusUnauthorized {
		t.Fatalf("旧密码应被拒绝, got %d", code)
	}
}

func TestChangePasswordWrongOld(t *testing.T) {
	h, uid := newAuthEnv(t)
	rec := changePasswordReq(h, uid, `{"oldPassword":"wrong-old-pass","newPassword":"newpassword99"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("旧密码错误应返回 401, got %d", rec.Code)
	}
	if code := loginReq(h, "admin", "password123456"); code != http.StatusOK {
		t.Fatalf("原密码不应被改动, got %d", code)
	}
}

func TestChangePasswordShortNew(t *testing.T) {
	h, uid := newAuthEnv(t)
	rec := changePasswordReq(h, uid, `{"oldPassword":"password123456","newPassword":"short7"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("新密码过短应返回 400, got %d", rec.Code)
	}
}
