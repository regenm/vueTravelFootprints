package handlers

import (
	"net/http"
	"regexp"
	"strings"
	"time"
	"travel-footprints/database"
	"travel-footprints/middleware"
	"travel-footprints/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var usernameRE = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

type AuthHandler struct {
	db        *database.DB
	jwtSecret string
}

func NewAuthHandler(db *database.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	user, err := h.db.GetUserByAccount(req.Account)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		writeError(w, http.StatusUnauthorized, "账号或密码不正确")
		return
	}

	token, err := h.signToken(*user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "登录凭证生成失败")
		return
	}

	writeOK(w, models.AuthResponse{Token: token, User: user.Public()})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := h.loadCurrentUser(w, r)
	if !ok {
		return
	}
	writeOK(w, user.Public())
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	user, ok := h.loadCurrentUser(w, r)
	if !ok {
		return
	}

	var req models.UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.DisplayName != nil {
		name := strings.TrimSpace(*req.DisplayName)
		if name == "" {
			name = user.Username
		}
		user.DisplayName = name
	}
	if req.Avatar != nil {
		avatar := strings.TrimSpace(*req.Avatar)
		if len(avatar) > 800 {
			writeError(w, http.StatusBadRequest, "头像地址过长")
			return
		}
		user.Avatar = avatar
	}
	if err := h.db.UpdateUser(user); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOK(w, user.Public())
}

func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}
	users, err := h.db.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	public := make([]models.UserPublic, 0, len(users))
	for _, u := range users {
		public = append(public, u.Public())
	}
	writeOK(w, public)
}

func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requireAdmin(w, r); !ok {
		return
	}

	var req models.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if !usernameRE.MatchString(req.Username) {
		writeError(w, http.StatusBadRequest, "用户名需为 3-20 位字母、数字或下划线")
		return
	}
	if !strings.Contains(req.Email, "@") || len(req.Email) < 5 {
		writeError(w, http.StatusBadRequest, "请输入有效邮箱")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "密码至少 8 位")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "无法创建账号")
		return
	}

	user := models.NewUser(req.Username, req.Email, string(hash), req.DisplayName)
	if req.Role == "admin" {
		user.Role = "admin"
	}

	if err := h.db.CreateUser(user); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeError(w, http.StatusConflict, "用户名或邮箱已被使用")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    user.Public(),
	})
}

func (h *AuthHandler) loadCurrentUser(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	uid := middleware.UserIDFromContext(r.Context())
	user, err := h.db.GetUserByID(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return nil, false
	}
	if user == nil {
		writeError(w, http.StatusUnauthorized, "用户不存在")
		return nil, false
	}
	return user, true
}

func (h *AuthHandler) requireAdmin(w http.ResponseWriter, r *http.Request) (*models.User, bool) {
	user, ok := h.loadCurrentUser(w, r)
	if !ok {
		return nil, false
	}
	if !user.IsAdmin() {
		writeError(w, http.StatusForbidden, "只有管理员可以执行此操作")
		return nil, false
	}
	return user, true
}

func (h *AuthHandler) signToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"uid":   user.ID,
		"uname": user.Username,
		"role":  user.Role,
		"exp":   time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
