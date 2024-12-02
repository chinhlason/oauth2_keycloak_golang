package server

import (
	"encoding/json"
	"net/http"
)

type IHandler interface {
	GetAdminTokenHandler(res http.ResponseWriter, req *http.Request)
	CreateRealms(res http.ResponseWriter, req *http.Request)
	CreateUser(res http.ResponseWriter, req *http.Request)
	SetPassword(res http.ResponseWriter, req *http.Request)
	GetUserToken(res http.ResponseWriter, req *http.Request)
	IntrospectToken(res http.ResponseWriter, req *http.Request)
	RefreshToken(res http.ResponseWriter, req *http.Request)
	UpdateUserInfo(res http.ResponseWriter, req *http.Request)
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

func (h *Handler) CreateRealms(res http.ResponseWriter, req *http.Request) {
	realms := []string{"realm1", "realm2", "realm3"}
	err := h.keyCloakClient.CreateRealms(realms)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("realms created"))
}

func (h *Handler) CreateUser(res http.ResponseWriter, req *http.Request) {
	user := &CreateUserRequest{
		Username:        "someday2",
		Email:           "hellothere2@gmail.com",
		FirstName:       "donuts",
		LastName:        "goodbye",
		Enabled:         true,
		EmailVerified:   false,
		RequiredActions: make([]string, 0),
		Attributes: struct {
			PhoneNumber string `json:"phoneNumber"`
			Locale      string `json:"locale"`
		}(struct {
			PhoneNumber string
			Locale      string
		}{
			PhoneNumber: "123456789",
			Locale:      "en",
		}),
	}

	err := h.keyCloakClient.CreateUser("users", user, "password")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("user created"))
}

func (h *Handler) SetPassword(res http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	password := req.URL.Query().Get("password")

	err := h.keyCloakClient.SetPassword(id, password)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("password set"))
}

func (h *Handler) GetUserToken(res http.ResponseWriter, req *http.Request) {
	username := req.URL.Query().Get("username")
	password := req.URL.Query().Get("password")

	result, err := h.keyCloakClient.GetUserToken(username, password)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	responseJson, _ := json.Marshal(result)

	res.WriteHeader(http.StatusOK)
	res.Write(responseJson)
}

func (h *Handler) IntrospectToken(res http.ResponseWriter, req *http.Request) {
	token := req.URL.Query().Get("token")

	result, err := h.keyCloakClient.IntrospectToken(token)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	responseJson, _ := json.Marshal(result)

	res.WriteHeader(http.StatusOK)
	res.Write(responseJson)
}

func (h *Handler) RefreshToken(res http.ResponseWriter, req *http.Request) {
	token := req.URL.Query().Get("token")
	response, err := h.keyCloakClient.RefreshToken(token)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resJson, err := json.Marshal(response)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte(resJson))
}

func (h *Handler) UpdateUserInfo(res http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	user := &UpdateUserInfo{
		FirstName: "updated",
		LastName:  "updated",
		Email:     "2222@gmail.com",
		Enabled:   true,
	}
	if err := h.keyCloakClient.UpdateUserInfo(id, user); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusOK)
	res.Write([]byte("user updated"))
}
