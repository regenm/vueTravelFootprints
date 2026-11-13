package handlers

import (
	"archive/zip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"travel-footprints/database"
	"travel-footprints/middleware"
)

type BackupHandler struct {
	db        *database.DB
	uploadDir string
}

func NewBackupHandler(db *database.DB, uploadDir string) *BackupHandler {
	return &BackupHandler{db: db, uploadDir: uploadDir}
}

// Export 打包数据库快照与全部上传图片为 zip 下载（仅管理员）。
// zip 内目录结构与运行目录一致：data/travel.db + uploads/...
func (h *BackupHandler) Export(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	user, err := h.db.GetUserByID(uid)
	if err != nil || user == nil || !user.IsAdmin() {
		writeError(w, http.StatusForbidden, "只有管理员可以执行此操作")
		return
	}

	// 先把完整 zip 写入临时文件再发送，出错时还能返回干净的 JSON 错误
	snapPath := tempFile(w, "travel-snapshot-*.db")
	if snapPath == "" {
		return
	}
	defer os.Remove(snapPath)
	if err := h.db.SnapshotTo(snapPath); err != nil {
		writeError(w, http.StatusInternalServerError, "生成数据库快照失败: "+err.Error())
		return
	}

	zipPath := tempFile(w, "travel-backup-*.zip")
	if zipPath == "" {
		return
	}
	defer os.Remove(zipPath)
	if err := buildZip(zipPath, snapPath, h.uploadDir); err != nil {
		writeError(w, http.StatusInternalServerError, "打包备份失败: "+err.Error())
		return
	}

	f, err := os.Open(zipPath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "读取备份失败: "+err.Error())
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", `attachment; filename="travel-backup-`+time.Now().Format("20060102-150405")+`.zip"`)
	_, _ = io.Copy(w, f)
}

func tempFile(w http.ResponseWriter, pattern string) string {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "创建临时文件失败")
		return ""
	}
	name := f.Name()
	f.Close()
	return name
}

func buildZip(zipPath, snapPath, uploadDir string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	if err := addFileToZip(zw, snapPath, "data/travel.db"); err != nil {
		zw.Close()
		return err
	}
	if info, err := os.Stat(uploadDir); err == nil && info.IsDir() {
		err = filepath.WalkDir(uploadDir, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, err := filepath.Rel(uploadDir, path)
			if err != nil {
				return err
			}
			return addFileToZip(zw, path, "uploads/"+filepath.ToSlash(rel))
		})
		if err != nil {
			zw.Close()
			return err
		}
	}
	return zw.Close()
}

func addFileToZip(zw *zip.Writer, path, name string) error {
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, src)
	return err
}
