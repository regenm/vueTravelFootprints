package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"travel-footprints/middleware"
	"travel-footprints/models"
)

type UploadHandler struct {
	uploadDir string
	baseURL   string
}

func NewUploadHandler(uploadDir, baseURL string) *UploadHandler {
	_ = os.MkdirAll(uploadDir, 0755)
	return &UploadHandler{uploadDir: uploadDir, baseURL: baseURL}
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	if uid == "" {
		writeError(w, http.StatusUnauthorized, "请先登录")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ImageUploadResponse{Success: false, Message: "请选择图片文件"})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ImageUploadResponse{Success: false, Message: "请选择图片文件"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		writeJSON(w, http.StatusBadRequest, models.ImageUploadResponse{Success: false, Message: "不支持的图片格式，仅支持 jpg/png/gif/webp"})
		return
	}

	filename := fmt.Sprintf("%s_%d%s", uid, time.Now().UnixNano(), ext)
	filePath := filepath.Join(h.uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ImageUploadResponse{Success: false, Message: "文件创建失败"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ImageUploadResponse{Success: false, Message: "文件保存失败"})
		return
	}

	writeJSON(w, http.StatusOK, models.ImageUploadResponse{
		Success: true,
		URL:     publicUploadURL(h.baseURL, filename),
	})
}

func publicUploadURL(baseURL, filename string) string {
	path := "/uploads/" + filename
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		return path
	}
	return base + path
}
