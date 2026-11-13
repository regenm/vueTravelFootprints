package models

type Marker struct {
	ID           string      `json:"id"`
	UserID       string      `json:"userId"`
	Name         string      `json:"name"`
	Longitude    string      `json:"longitude"`
	Latitude     string      `json:"latitude"`
	Address      string      `json:"address"`
	Photos       []string    `json:"photos"`
	Category     string      `json:"category"`
	Notes        string      `json:"notes"`
	VisitDate    string      `json:"visitDate"`
	ProvinceCode string      `json:"provinceCode"`
	CityCode     string      `json:"cityCode"`
	ProvinceName string      `json:"provinceName"`
	CityName     string      `json:"cityName"`
	DistrictCode string      `json:"districtCode"`
	DistrictName string      `json:"districtName"`
	TripID       string      `json:"tripId"`
	IsPublic     bool        `json:"isPublic"`
	CreatedAt    string      `json:"createdAt"`
	UpdatedAt    string      `json:"updatedAt"`
	Author       *UserPublic `json:"author,omitempty"`
}

// RegionInfo 足迹的行政区划归属
type RegionInfo struct {
	ProvinceCode string `json:"provinceCode"`
	CityCode     string `json:"cityCode"`
	ProvinceName string `json:"provinceName"`
	CityName     string `json:"cityName"`
	DistrictCode string `json:"districtCode"`
	DistrictName string `json:"districtName"`
}

type CreateMarkerRequest struct {
	Name      string   `json:"name"`
	Longitude string   `json:"longitude"`
	Latitude  string   `json:"latitude"`
	Address   string   `json:"address"`
	Photos    []string `json:"photos"`
	Category  string   `json:"category"`
	Notes     string   `json:"notes"`
	VisitDate string   `json:"visitDate"`
	ShareID   string   `json:"shareId"`
	TripID    string   `json:"tripId"`
}

type UpdateMarkerRequest struct {
	Name      *string  `json:"name"`
	Longitude *string  `json:"longitude"`
	Latitude  *string  `json:"latitude"`
	Address   *string  `json:"address"`
	Photos    []string `json:"photos"`
	Category  *string  `json:"category"`
	Notes     *string  `json:"notes"`
	VisitDate *string  `json:"visitDate"`
	TripID    *string  `json:"tripId"`
}

func NewMarker(userID string, req CreateMarkerRequest) Marker {
	now := NowISO()
	photos := req.Photos
	if photos == nil {
		photos = []string{}
	}
	return Marker{
		ID:        generateID(),
		UserID:    userID,
		Name:      req.Name,
		Longitude: req.Longitude,
		Latitude:  req.Latitude,
		Address:   req.Address,
		Photos:    photos,
		Category:  req.Category,
		Notes:     req.Notes,
		VisitDate: req.VisitDate,
		TripID:    req.TripID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
