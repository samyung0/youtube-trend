package handler

import (
	"encoding/json"
	"net/http"

	internalAuth "github.com/thumbtrend/backend/internal/auth"
	"github.com/thumbtrend/backend/internal/model"
	"github.com/thumbtrend/backend/internal/store"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	oauthCfg  *oauth2.Config
	userStore *store.UserStore
	jwtSecret string
	frontURL  string
}

func NewAuthHandler(oauthCfg *oauth2.Config, us *store.UserStore, jwtSecret, frontURL string) *AuthHandler {
	return &AuthHandler{oauthCfg: oauthCfg, userStore: us, jwtSecret: jwtSecret, frontURL: frontURL}
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	url := h.oauthCfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "missing code")
		return
	}

	gUser, err := internalAuth.GetGoogleUser(r.Context(), h.oauthCfg, code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get google user")
		return
	}

	user, err := h.userStore.FindByGoogleID(r.Context(), gUser.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to find user")
		return
	}
	if user == nil {
		user = &model.User{
			GoogleID:  gUser.ID,
			Email:     gUser.Email,
			Name:      gUser.Name,
			AvatarURL: gUser.Picture,
		}
		if err := h.userStore.Create(r.Context(), user); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create user")
			return
		}
	}

	token, err := internalAuth.GenerateToken(user.ID, user.Email, h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	http.Redirect(w, r, h.frontURL+"/auth/callback?token="+token, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := internalAuth.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, err := h.userStore.FindByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	sub, _ := h.userStore.GetSubscription(r.Context(), userID)
	resp := map[string]interface{}{
		"user":   user,
		"is_pro": sub != nil && sub.IsActive(),
	}

	b, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}
