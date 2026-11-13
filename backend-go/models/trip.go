package models

type Trip struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"createdAt"`
}

type CreateTripRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
	Notes     string `json:"notes"`
}

type UpdateTripRequest struct {
	Name      *string `json:"name"`
	StartDate *string `json:"startDate"`
	EndDate   *string `json:"endDate"`
	Notes     *string `json:"notes"`
}

func NewTrip(userID string, req CreateTripRequest) Trip {
	return Trip{
		ID:        generateID(),
		UserID:    userID,
		Name:      req.Name,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Notes:     req.Notes,
		CreatedAt: NowISO(),
	}
}
