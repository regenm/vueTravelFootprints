package handlers

import (
	"encoding/json"
	"net/http"
	"travel-footprints/database"
	"travel-footprints/models"
)

type MarkerHandler struct {
	db *database.DB
}

func NewMarkerHandler(db *database.DB) *MarkerHandler {
	return &MarkerHandler{db: db}
}

func (h *MarkerHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	markers, err := h.db.GetAllMarkers()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    markers,
	})
}

func (h *MarkerHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	marker, err := h.db.GetMarkerByID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if marker == nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "足迹不存在",
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    marker,
	})
}

func (h *MarkerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMarkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	if req.Name == "" || req.Longitude == "" || req.Latitude == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "名称或坐标不能为空",
		})
		return
	}

	marker := models.NewMarker(req)
	if err := h.db.CreateMarker(marker); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	markers, _ := h.db.GetAllMarkers()
	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    markers,
	})
}

func (h *MarkerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req models.UpdateMarkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "请求格式错误",
		})
		return
	}

	marker, err := h.db.UpdateMarker(id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}
	if marker == nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "足迹不存在",
		})
		return
	}

	markers, _ := h.db.GetAllMarkers()
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    markers,
	})
}

func (h *MarkerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.db.DeleteMarker(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	markers, _ := h.db.GetAllMarkers()
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    markers,
	})
}

func (h *MarkerHandler) Search(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	category := r.URL.Query().Get("category")
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	markers, err := h.db.SearchMarkers(keyword, category, startDate, endDate)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    markers,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}