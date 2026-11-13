package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"travel-footprints/database"
	"travel-footprints/middleware"
	"travel-footprints/models"
)

type MarkerHandler struct {
	db        *database.DB
	geocode   *GeocodeClient
	uploadDir string
}

func NewMarkerHandler(db *database.DB, geocode *GeocodeClient, uploadDir string) *MarkerHandler {
	return &MarkerHandler{db: db, geocode: geocode, uploadDir: uploadDir}
}

// removeUnreferencedPhotos 删除不再被任何足迹引用的上传图片文件
func (h *MarkerHandler) removeUnreferencedPhotos(photos []string, excludeMarkerID string) {
	for _, url := range photos {
		name := photoFilename(url)
		if name == "" {
			continue
		}
		if n, err := h.db.CountMarkersReferencingPhoto(url, excludeMarkerID); err != nil {
			log.Printf("查询图片引用失败 %s: %v", url, err)
			continue
		} else if n > 0 {
			continue
		}
		if err := os.Remove(filepath.Join(h.uploadDir, name)); err != nil && !os.IsNotExist(err) {
			log.Printf("清理图片文件失败 %s: %v", name, err)
		}
	}
}

// photoFilename 从图片 URL 提取 uploads 目录下的文件名，非法路径返回空串
func photoFilename(url string) string {
	i := strings.LastIndex(url, "/uploads/")
	if i < 0 {
		return ""
	}
	name := url[i+len("/uploads/"):]
	if name == "" || strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return ""
	}
	return name
}

// subtractPhotos 返回 old 中有、new 中没有的图片 URL
func subtractPhotos(old, cur []string) []string {
	kept := make(map[string]bool, len(cur))
	for _, u := range cur {
		kept[u] = true
	}
	var removed []string
	for _, u := range old {
		if !kept[u] {
			removed = append(removed, u)
		}
	}
	return removed
}

func (h *MarkerHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	markers, err := h.db.GetMarkersByUser(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOK(w, h.db.AttachAuthors(markers))
}

func (h *MarkerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	marker, err := h.db.GetMarkerByID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if marker == nil {
		writeError(w, http.StatusNotFound, "足迹不存在")
		return
	}
	if marker.UserID != uid && !h.canAccessMarker(uid, marker) {
		writeError(w, http.StatusForbidden, "没有权限查看该足迹")
		return
	}
	writeOK(w, withAuthor(h.db, *marker))
}

func (h *MarkerHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	var req models.CreateMarkerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.Name == "" || req.Longitude == "" || req.Latitude == "" {
		writeError(w, http.StatusBadRequest, "名称或坐标不能为空")
		return
	}

	if req.ShareID != "" {
		if ok, err := h.canEditShare(uid, req.ShareID); err != nil || !ok {
			writeError(w, http.StatusForbidden, "没有权限在此分享中添加足迹")
			return
		}
	}

	marker := models.NewMarker(uid, req)
	if region := h.geocode.ReverseGeocode(marker.Longitude, marker.Latitude); region != nil {
		marker.ProvinceCode = region.ProvinceCode
		marker.CityCode = region.CityCode
		marker.ProvinceName = region.ProvinceName
		marker.CityName = region.CityName
		marker.DistrictCode = region.DistrictCode
		marker.DistrictName = region.DistrictName
	}
	if err := h.db.CreateMarker(marker); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if req.ShareID != "" {
		_ = h.db.AddShareMarker(req.ShareID, marker.ID)
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: withAuthor(h.db, marker)})
}

func (h *MarkerHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	existing, err := h.db.GetMarkerByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "足迹不存在")
		return
	}
	if existing.UserID != uid {
		writeError(w, http.StatusForbidden, "只能编辑自己的足迹")
		return
	}

	var req models.UpdateMarkerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	coordChanged := (req.Longitude != nil && *req.Longitude != existing.Longitude) ||
		(req.Latitude != nil && *req.Latitude != existing.Latitude)
	// 区县码为空说明是区县级点亮上线前的旧数据，编辑保存时一并反查补齐
	needRegion := coordChanged || existing.ProvinceCode == "" || existing.DistrictCode == ""
	var region *models.RegionInfo
	if needRegion {
		nextLng, nextLat := existing.Longitude, existing.Latitude
		if req.Longitude != nil {
			nextLng = *req.Longitude
		}
		if req.Latitude != nil {
			nextLat = *req.Latitude
		}
		region = h.geocode.ReverseGeocode(nextLng, nextLat)
	}

	marker, err := h.db.UpdateMarker(id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.removeUnreferencedPhotos(subtractPhotos(existing.Photos, marker.Photos), marker.ID)
	if region != nil {
		if err := h.db.UpdateMarkerRegion(marker.ID, *region); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		marker.ProvinceCode = region.ProvinceCode
		marker.CityCode = region.CityCode
		marker.ProvinceName = region.ProvinceName
		marker.CityName = region.CityName
		marker.DistrictCode = region.DistrictCode
		marker.DistrictName = region.DistrictName
	}
	writeOK(w, withAuthor(h.db, *marker))
}

func (h *MarkerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	id := r.PathValue("id")

	existing, err := h.db.GetMarkerByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "足迹不存在")
		return
	}
	if existing.UserID != uid {
		writeError(w, http.StatusForbidden, "只能删除自己的足迹")
		return
	}

	if err := h.db.DeleteMarker(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.removeUnreferencedPhotos(existing.Photos, id)
	writeOK(w, map[string]string{"id": id})
}

func (h *MarkerHandler) Search(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	markers, err := h.db.SearchMarkers(
		uid,
		r.URL.Query().Get("keyword"),
		r.URL.Query().Get("category"),
		r.URL.Query().Get("startDate"),
		r.URL.Query().Get("endDate"),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOK(w, h.db.AttachAuthors(markers))
}

func withAuthor(db *database.DB, marker models.Marker) models.Marker {
	list := db.AttachAuthors([]models.Marker{marker})
	return list[0]
}

func (h *MarkerHandler) canAccessMarker(uid string, marker *models.Marker) bool {
	shares, err := h.db.ListSharesForMember(uid)
	if err != nil {
		return false
	}
	owned, _ := h.db.ListSharesByOwner(uid)
	shares = append(shares, owned...)
	for i := range shares {
		markers, err := h.db.GetShareMarkers(&shares[i])
		if err != nil {
			continue
		}
		for _, m := range markers {
			if m.ID == marker.ID {
				return true
			}
		}
	}
	return false
}

func (h *MarkerHandler) canEditShare(uid, shareID string) (bool, error) {
	share, err := h.db.GetShareByID(shareID)
	if err != nil || share == nil {
		return false, err
	}
	if share.OwnerID == uid {
		return true, nil
	}
	perm, err := h.db.GetMemberPermission(shareID, uid)
	if err != nil {
		return false, err
	}
	return perm == "edit", nil
}
