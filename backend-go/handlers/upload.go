package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"travel-footprints/models"
)

type UploadHandler struct {
	uploadDir string
	baseURL   string
}

func NewUploadHandler(uploadDir, baseURL string) *UploadHandler {
	os.MkdirAll(uploadDir, 0755)
	return &UploadHandler{
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)

	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ImageUploadResponse{
			Success: false,
			Message: "请选择图片文件",
		})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}

	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		writeJSON(w, http.StatusBadRequest, models.ImageUploadResponse{
			Success: false,
			Message: "不支持的图片格式，仅支持 jpg/png/gif/webp",
		})
		return
	}

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), randomStr(8), ext)
	filePath := filepath.Join(h.uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ImageUploadResponse{
			Success: false,
			Message: "文件创建失败",
		})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ImageUploadResponse{
			Success: false,
			Message: "文件保存失败",
		})
		return
	}

	url := fmt.Sprintf("%s/uploads/%s", h.baseURL, filename)

	writeJSON(w, http.StatusOK, models.ImageUploadResponse{
		Success: true,
		URL:     url,
	})
}

func randomStr(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}