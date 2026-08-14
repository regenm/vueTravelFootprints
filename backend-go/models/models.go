package models

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Marker struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Longitude string   `json:"longitude"`
	Latitude  string   `json:"latitude"`
	Photos    []string `json:"photos"`
	Category  string   `json:"category"`
	Notes     string   `json:"notes"`
	VisitDate string   `json:"visitDate"`
	CreatedAt string   `json:"createdAt"`
	UpdatedAt string   `json:"updatedAt"`
}

type CreateMarkerRequest struct {
	Name      string   `json:"name"`
	Longitude string   `json:"longitude"`
	Latitude  string   `json:"latitude"`
	Photos    []string `json:"photos"`
	Category  string   `json:"category"`
	Notes     string   `json:"notes"`
	VisitDate string   `json:"visitDate"`
}

type UpdateMarkerRequest struct {
	Name      *string  `json:"name"`
	Longitude *string  `json:"longitude"`
	Latitude  *string  `json:"latitude"`
	Photos    []string `json:"photos"`
	Category  *string  `json:"category"`
	Notes     *string  `json:"notes"`
	VisitDate *string  `json:"visitDate"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ImageUploadResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewMarker(req CreateMarkerRequest) Marker {
	now := time.Now().UTC().Format(time.RFC3339)
	return Marker{
		ID:        generateID(),
		Name:      req.Name,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
		Photos:    req.Photos,
		Category:  req.Category,
		Notes:     req.Notes,
		VisitDate: req.VisitDate,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}