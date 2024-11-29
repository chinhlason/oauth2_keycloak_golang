package server

import (
	"encoding/json"
	"net/http"
)

type IHandler interface {
	GetAdminTokenHandler(res http.ResponseWriter, req *http.Request)
}

type Handler struct {
	keyCloakClient IKeyCloak
}

func NewHandler(keyCloakClient IKeyCloak) IHandler {
	return &Handler{keyCloakClient}
}

func (h *Handler) GetAdminTokenHandler(res http.ResponseWriter, req *http.Request) {
	result, err := h.keyCloakClient.GetAdminToken()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	responseJson, _ := json.Marshal(result)

	res.WriteHeader(http.StatusOK)
	res.Write(responseJson)
}
