package server

import (
	"encoding/json"
	"net/http"
)

type IHandler interface {
	LoginWithGoogle(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
}

type Handler struct {
	service IService
}

func NewHandler(service IService) IHandler {
	return &Handler{service}
}

func (h *Handler) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	tokenFromCode, err := GetTokenFromCode(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token, err := h.service.LoginWithGoogle(tokenFromCode.AccessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	responseJson, _ := json.Marshal(token)
	w.WriteHeader(http.StatusOK)
	w.Write(responseJson)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.service.Logout(refreshToken.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logout success"))
}
