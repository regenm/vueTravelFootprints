package models

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	DisplayName  string `json:"displayName"`
	Avatar       string `json:"avatar"`
	Role         string `json:"role"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type UserPublic struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
	Avatar      string `json:"avatar"`
	Role        string `json:"role"`
}

func (u User) Public() UserPublic {
	name := u.DisplayName
	if name == "" {
		name = u.Username
	}
	role := u.Role
	if role == "" {
		role = "user"
	}
	return UserPublic{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: name,
		Avatar:      u.Avatar,
		Role:        role,
	}
}

func (u User) IsAdmin() bool {
	return u.Role == "admin"
}

type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type LoginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"displayName"`
	Avatar      *string `json:"avatar"`
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  UserPublic `json:"user"`
}

func NewUser(username, email, passwordHash, displayName string) User {
	now := NowISO()
	if displayName == "" {
		displayName = username
	}
	return User{
		ID:           generateID(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
