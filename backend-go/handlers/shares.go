package handlers

import (
	"database/sql"
	"net/http"
	"strings"
	"time"
	"travel-footprints/database"
	"travel-footprints/middleware"
	"travel-footprints/models"
)

type ShareHandler struct {
	db *database.DB
}

func NewShareHandler(db *database.DB) *ShareHandler {
	return &ShareHandler{db: db}
}

func (h *ShareHandler) Create(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	var req models.CreateShareRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.Type == "selected" && len(req.MarkerIDs) == 0 {
		writeError(w, http.StatusBadRequest, "请至少选择一条足迹")
		return
	}

	for _, mid := range req.MarkerIDs {
		m, err := h.db.GetMarkerByID(mid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		if m == nil || m.UserID != uid {
			writeError(w, http.StatusForbidden, "只能分享自己的足迹")
			return
		}
	}

	share := models.NewShare(uid, req)
	if err := h.db.CreateShare(share, req.MarkerIDs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	summary, err := h.buildSummary(&share, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Data: summary})
}

func (h *ShareHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	shares, err := h.db.ListSharesByOwner(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	list := make([]models.ShareSummary, 0, len(shares))
	for i := range shares {
		s, err := h.buildSummary(&shares[i], uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		list = append(list, *s)
	}
	writeOK(w, list)
}

func (h *ShareHandler) Inbox(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	shares, err := h.db.ListSharesForMember(uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	list := make([]models.ShareSummary, 0, len(shares))
	for i := range shares {
		s, err := h.buildSummary(&shares[i], uid)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		list = append(list, *s)
	}
	writeOK(w, list)
}

func (h *ShareHandler) Update(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	share, err := h.db.GetShareByID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if share == nil || share.OwnerID != uid {
		writeError(w, http.StatusForbidden, "只有创建者可以修改分享")
		return
	}

	var req models.UpdateShareRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			title = "我的旅行足迹"
		}
		share.Title = title
	}
	if req.Description != nil {
		share.Description = strings.TrimSpace(*req.Description)
	}
	if req.Permission != nil {
		perm := *req.Permission
		if perm != "edit" {
			perm = "view"
		}
		share.Permission = perm
	}
	if req.IsPublic != nil {
		share.IsPublic = *req.IsPublic
	}
	if err := h.db.UpdateShare(share); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	summary, err := h.buildSummary(share, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOK(w, summary)
}

func (h *ShareHandler) Delete(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	if err := h.db.DeleteShare(r.PathValue("id"), uid); err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "分享不存在或无权删除")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOK(w, map[string]string{"id": r.PathValue("id")})
}

func (h *ShareHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	share, err := h.db.GetShareByID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if share == nil || share.OwnerID != uid {
		writeError(w, http.StatusForbidden, "只有创建者可以邀请成员")
		return
	}

	var req models.AddMemberRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式错误")
		return
	}
	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		writeError(w, http.StatusNotFound, "未找到该用户名")
		return
	}
	if user.ID == uid {
		writeError(w, http.StatusBadRequest, "不能邀请自己")
		return
	}

	perm := req.Permission
	if perm != "edit" {
		perm = "view"
	}
	if err := h.db.AddShareMember(share.ID, user.ID, perm); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	summary, err := h.buildSummary(share, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOK(w, summary)
}

func (h *ShareHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	uid := middleware.UserIDFromContext(r.Context())
	share, err := h.db.GetShareByID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if share == nil {
		writeError(w, http.StatusNotFound, "分享不存在")
		return
	}

	targetID := r.PathValue("userId")
	if targetID == "me" {
		targetID = uid
	}
	if share.OwnerID != uid && targetID != uid {
		writeError(w, http.StatusForbidden, "只有创建者可以移除成员")
		return
	}
	if targetID == share.OwnerID {
		writeError(w, http.StatusBadRequest, "不能移除地图创建者")
		return
	}
	if err := h.db.RemoveShareMember(share.ID, targetID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOK(w, map[string]string{"ok": "1"})
}

func (h *ShareHandler) PublicGet(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	share, err := h.db.GetShareByToken(token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if share == nil {
		writeError(w, http.StatusNotFound, "分享不存在或已失效")
		return
	}
	if share.ExpiresAt != "" {
		if exp, err := time.Parse(time.RFC3339, share.ExpiresAt); err == nil && time.Now().After(exp) {
			writeError(w, http.StatusGone, "分享链接已过期")
			return
		}
	}

	uid := middleware.UserIDFromContext(r.Context())
	canView, canEdit := h.resolveAccess(share, uid)
	if !canView {
		writeError(w, http.StatusForbidden, "该分享为私密链接，请登录后访问")
		return
	}

	markers, err := h.db.GetShareMarkers(share)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	markers = h.db.AttachAuthors(markers)

	summary, err := h.buildSummary(share, uid)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	summary.MyPermission = "view"
	if canEdit {
		summary.MyPermission = "edit"
	}

	writeOK(w, models.ShareView{
		Share:   *summary,
		Markers: markers,
		CanEdit: canEdit,
	})
}

func (h *ShareHandler) resolveAccess(share *models.Share, uid string) (canView, canEdit bool) {
	if share.IsPublic {
		canView = true
	}
	if uid == "" {
		return canView, false
	}
	if share.OwnerID == uid {
		return true, true
	}
	perm, err := h.db.GetMemberPermission(share.ID, uid)
	if err != nil || perm == "" {
		return canView, false
	}
	return true, perm == "edit"
}

func (h *ShareHandler) buildSummary(share *models.Share, viewerID string) (*models.ShareSummary, error) {
	owner, err := h.db.GetUserByID(share.OwnerID)
	if err != nil {
		return nil, err
	}
	ownerPub := models.UserPublic{}
	if owner != nil {
		ownerPub = owner.Public()
	}
	count, err := h.db.CountShareMarkers(share)
	if err != nil {
		return nil, err
	}
	members, err := h.db.GetShareMembers(share.ID)
	if err != nil {
		return nil, err
	}
	if members == nil {
		members = []models.ShareMember{}
	}
	participants := make([]models.UserPublic, 0, 1+len(members))
	seen := map[string]bool{}
	if ownerPub.ID != "" {
		participants = append(participants, ownerPub)
		seen[ownerPub.ID] = true
	}
	for _, m := range members {
		if seen[m.UserID] {
			continue
		}
		seen[m.UserID] = true
		participants = append(participants, models.UserPublic{
			ID:          m.UserID,
			Username:    m.Username,
			DisplayName: m.DisplayName,
			Avatar:      m.Avatar,
			Role:        "user",
		})
	}
	perm := "view"
	if viewerID == share.OwnerID {
		perm = "edit"
	} else if viewerID != "" {
		if p, err := h.db.GetMemberPermission(share.ID, viewerID); err == nil && p != "" {
			perm = p
		}
	}
	return &models.ShareSummary{
		Share:        *share,
		Owner:        ownerPub,
		MarkerCount:  count,
		Members:      members,
		Participants: participants,
		MyPermission: perm,
	}, nil
}
