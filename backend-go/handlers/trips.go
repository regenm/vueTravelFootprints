package handlers

import (
	"net/http"
	"travel-footprints/database"
	"travel-footprints/middleware"
	"travel-footprints/models"
)

type TripHandler struct {
	db *database.DB
}

func NewTripHandler(db *database.DB) *TripHandler {
	return &TripHandler{db: db}
}

func (h *TripHandler) List(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	trips, err := h.db.GetTripsByUser(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOK(w, trips)
}

func (h *TripHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	var req models.CreateTripRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "行程名称不能为空")
		return
	}
	trip := models.NewTrip(uid, req)
	if err := h.db.CreateTrip(trip); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: trip})
}

func (h *TripHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	var req models.UpdateTripRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	trip, err := h.db.UpdateTrip(r.PathValue("id"), uid, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if trip == nil {
		writeError(w, http.StatusNotFound, "行程不存在")
		return
	}
	writeOK(w, trip)
}

func (h *TripHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	if err := h.db.DeleteTrip(r.PathValue("id"), uid); err != nil {
		writeError(w, http.StatusNotFound, "行程不存在")
		return
	}
	writeOK(w, map[string]string{"id": r.PathValue("id")})
}
