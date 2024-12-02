package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	ADMIN_ACCESS_TOKEN  = "admin_access_token"
	ADMIN_REFRESH_TOKEN = "admin_refresh_token"
)

type IKeyCloak interface {
	GetAdminToken() (*TokenResponse, error)
	CreateRealms(realms []string) error
	CreateUser(realm string, req *CreateUserRequest, password string) (string, error)
	SetPassword(id string, password string) error
	GetUserToken(username, password string) (*TokenResponse, error)
	IntrospectToken(token string) (*UserInfoResponse, error)
	RefreshToken(refreshToken string) (*TokenResponse, error)
	UpdateUserInfo(id string, req *UpdateUserInfo) error
	GetUserByEmail(email string) (*UserResponse, error)
}

type KeyCloakClient struct {
	client *http.Client
	addr   string

	rdb *redis.Client
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func NewKeyCloakClient(addr string, timeout time.Duration, rdb *redis.Client) IKeyCloak {
	httpClient := &http.Client{Timeout: timeout}
	return &KeyCloakClient{
		client: httpClient,
		addr:   addr,
		rdb:    rdb,
	}
}

func (c *KeyCloakClient) GetAdminToken() (*TokenResponse, error) {
	var result TokenResponse
	accessToken, err := c.rdb.Get(context.Background(), ADMIN_ACCESS_TOKEN).Result()
	if err == nil {
		result.AccessToken = accessToken
	}

	refreshToken, err := c.rdb.Get(context.Background(), ADMIN_REFRESH_TOKEN).Result()
	if err == nil {
		result.RefreshToken = refreshToken
	}

	if result.AccessToken != "" && result.RefreshToken != "" {
		return &result, nil
	}

	urlStr := fmt.Sprintf("%s/realms/master/protocol/openid-connect/token", c.addr)

	form := url.Values{}
	form.Add("client_id", os.Getenv("KEYCLOAK_ADMIN_CLIENT_ID"))
	form.Add("username", os.Getenv("KEYCLOAK_ADMIN_USERNAME"))
	form.Add("password", os.Getenv("KEYCLOAK_ADMIN_PASSWORD"))
	form.Add("grant_type", "password")

	request, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return &TokenResponse{}, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.client.Do(request)
	if err != nil {
		return &TokenResponse{}, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&result)

	err = c.rdb.Set(
		context.Background(),
		ADMIN_ACCESS_TOKEN,
		result.AccessToken,
		time.Duration(result.ExpiredIn-1)*time.Second,
	).Err()
	if err != nil {
		return &TokenResponse{}, err
	}

	err = c.rdb.Set(
		context.Background(),
		ADMIN_REFRESH_TOKEN,
		result.RefreshToken,
		time.Duration(result.RefreshTokenExpiredIn-1)*time.Second,
	).Err()
	if err != nil {
		return &TokenResponse{}, err
	}

	return &result, nil
}

func (c *KeyCloakClient) CreateRealms(realms []string) error {
	token, err := c.GetAdminToken()
	if err != nil {
		return err
	}

	urlStr := fmt.Sprintf("%s/admin/realms", c.addr)

	for _, realm := range realms {
		bodyJson, err := json.Marshal(map[string]interface{}{
			"realm":               realm,
			"enabled":             true,
			"sslRequired":         "external",
			"accessTokenLifespan": os.Getenv("ACCESS_TOKEN_EXPIRATION"),
			"displayName":         realm,
		})
		if err != nil {
			return err
		}
		request, err := http.NewRequest("POST", urlStr, strings.NewReader(string(bodyJson)))
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

		response, err := c.client.Do(request)
		if err != nil {
			return err
		}

		if response.StatusCode == http.StatusConflict {
			return errors.New("realm already created")
		}

		if response.StatusCode != http.StatusCreated {
			return errors.New("failed to create realms")
		}

		defer response.Body.Close()
	}
	return nil
}

func (c *KeyCloakClient) CreateUser(realm string, req *CreateUserRequest, password string) (string, error) {
	token, err := c.GetAdminToken()
	if err != nil {
		return "", err
	}

	if password == "default" && realm == "default" {
		password = os.Getenv("DEFAULT_PASSWORD")
		realm = os.Getenv("KEYCLOAK_REALM")
	}

	urlStr := fmt.Sprintf("%s/admin/realms/%s/users", c.addr, realm)

	bodyJson, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest("POST", urlStr, strings.NewReader(string(bodyJson)))
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := c.client.Do(request)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusCreated {
		return "", errors.New(string(body))
	}

	location := response.Header.Get("Location")
	if location == "" {
		return "", errors.New("failed to create user")
	}

	id := strings.Split(location, "/")[len(strings.Split(location, "/"))-1]
	if err = c.SetPassword(id, password); err != nil {
		return "", err
	}

	defer response.Body.Close()
	return id, nil
}

func (c *KeyCloakClient) SetPassword(id string, password string) error {
	token, err := c.GetAdminToken()
	if err != nil {
		return err
	}

	urlStr := fmt.Sprintf("%s/admin/realms/%s/users/%s/reset-password", c.addr, os.Getenv("KEYCLOAK_REALM"), id)

	bodyJson, err := json.Marshal(map[string]interface{}{
		"type":      "password",
		"value":     password,
		"temporary": false,
	})
	if err != nil {
		return err
	}
	request, err := http.NewRequest("PUT", urlStr, strings.NewReader(string(bodyJson)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		return errors.New("failed to set password")
	}

	defer response.Body.Close()
	return nil
}

func (c *KeyCloakClient) GetUserToken(username, password string) (*TokenResponse, error) {
	urlStr := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.addr, os.Getenv("KEYCLOAK_REALM"))

	if password == "default" {
		password = os.Getenv("DEFAULT_PASSWORD")
	}

	form := url.Values{}
	form.Add("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
	form.Add("username", username)
	form.Add("password", password)
	form.Add("grant_type", "password")
	form.Add("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET"))

	request, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return &TokenResponse{}, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.client.Do(request)
	if err != nil {
		return &TokenResponse{}, err
	}
	defer response.Body.Close()

	var result TokenResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return &TokenResponse{}, err
	}

	return &result, nil
}

func (c *KeyCloakClient) IntrospectToken(token string) (*UserInfoResponse, error) {
	urlStr := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token/introspect", c.addr, os.Getenv("KEYCLOAK_REALM"))

	form := url.Values{}
	form.Add("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
	form.Add("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET"))
	form.Add("token", token)

	request, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var result UserInfoResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	if !result.Active {
		return nil, errors.New("token is not active")
	}

	return &result, nil
}

func (c *KeyCloakClient) RefreshToken(refreshToken string) (*TokenResponse, error) {
	urlStr := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.addr, os.Getenv("KEYCLOAK_REALM"))

	form := url.Values{}
	form.Add("client_id", os.Getenv("KEYCLOAK_CLIENT_ID"))
	form.Add("client_secret", os.Getenv("KEYCLOAK_CLIENT_SECRET"))
	form.Add("refresh_token", refreshToken)
	form.Add("grant_type", "refresh_token")

	request, err := http.NewRequest("POST", urlStr, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(body))
	}

	var result TokenResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *KeyCloakClient) UpdateUserInfo(id string, req *UpdateUserInfo) error {
	token, err := c.GetAdminToken()
	if err != nil {
		return err
	}

	urlStr := fmt.Sprintf("%s/admin/realms/%s/users/%s", c.addr, os.Getenv("KEYCLOAK_REALM"), id)

	bodyJson, err := json.Marshal(req)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("PUT", urlStr, strings.NewReader(string(bodyJson)))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}

	defer response.Body.Close()
	return nil
}

func (c *KeyCloakClient) GetUserByEmail(email string) (*UserResponse, error) {
	token, err := c.GetAdminToken()
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf("%s/admin/realms/%s/users?email=%s", c.addr, os.Getenv("KEYCLOAK_REALM"), email)

	request, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(body))
	}

	var result []UserResponse
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	for _, user := range result {
		if user.Email == email {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}

type TokenResponse struct {
	AccessToken           string `json:"access_token"`
	ExpiredIn             int    `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiredIn int    `json:"refresh_expires_in"`
	SessionState          string `json:"session_state"`
	Scope                 string `json:"scope"`
}

type CreateUserRequest struct {
	Email           string   `json:"email"`
	Username        string   `json:"username"`
	FirstName       string   `json:"firstName"`
	LastName        string   `json:"lastName"`
	Groups          []string `json:"groups"`
	Enabled         bool     `json:"enabled"`
	EmailVerified   bool     `json:"emailVerified"`
	RequiredActions []string `json:"requiredActions"`
	Attributes      struct {
		PhoneNumber string `json:"phoneNumber"`
		Locale      string `json:"locale"`
	} `json:"attributes"`
}

type UserInfoResponse struct {
	GivenName    string `json:"given_name"`
	FamilyName   string `json:"family_name"`
	Email        string `json:"email"`
	ClientId     string `json:"client_id"`
	Username     string `json:"username"`
	TokenType    string `json:"token_type"`
	Active       bool   `json:"active"`
	Scope        string `json:"scope"`
	RealmsAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess struct {
		Account struct {
			Roles []string `json:"roles"`
		}
	} `json:"resource_access"`
}

type UpdateUserInfo struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Enabled   bool   `json:"enabled"`
}

type UserResponse struct {
	Id        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}
