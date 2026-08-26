package models

type Share struct {
	ID          string `json:"id"`
	OwnerID     string `json:"ownerId"`
	Token       string `json:"token"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Permission  string `json:"permission"`
	IsPublic    bool   `json:"isPublic"`
	ExpiresAt   string `json:"expiresAt"`
	CreatedAt   string `json:"createdAt"`
}

type ShareMember struct {
	UserID      string `json:"userId"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Permission  string `json:"permission"`
	CreatedAt   string `json:"createdAt"`
}

type ShareSummary struct {
	Share
	Owner          UserPublic    `json:"owner"`
	MarkerCount    int           `json:"markerCount"`
	Members        []ShareMember `json:"members"`
	Participants   []UserPublic  `json:"participants"`
	MyPermission   string        `json:"myPermission"`
}

type ShareView struct {
	Share    ShareSummary `json:"share"`
	Markers  []Marker     `json:"markers"`
	CanEdit  bool         `json:"canEdit"`
}

type CreateShareRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Permission  string   `json:"permission"`
	IsPublic    *bool    `json:"isPublic"`
	ExpiresAt   string   `json:"expiresAt"`
	MarkerIDs   []string `json:"markerIds"`
}

type AddMemberRequest struct {
	Username   string `json:"username"`
	Permission string `json:"permission"`
}

type UpdateShareRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Permission  *string `json:"permission"`
	IsPublic    *bool   `json:"isPublic"`
}

func NewShare(ownerID string, req CreateShareRequest) Share {
	shareType := req.Type
	if shareType != "selected" {
		shareType = "all"
	}
	perm := req.Permission
	if perm != "edit" {
		perm = "view"
	}
	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	title := req.Title
	if title == "" {
		title = "我的旅行足迹"
	}
	return Share{
		ID:          generateID(),
		OwnerID:     ownerID,
		Token:       generateToken(),
		Title:       title,
		Description: req.Description,
		Type:        shareType,
		Permission:  perm,
		IsPublic:    isPublic,
		ExpiresAt:   req.ExpiresAt,
		CreatedAt:   NowISO(),
	}
}
