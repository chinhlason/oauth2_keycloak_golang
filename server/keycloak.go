package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
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
	GetAdminToken() (*AdminTokenResponse, error)
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

func (c *KeyCloakClient) GetAdminToken() (*AdminTokenResponse, error) {
	var result AdminTokenResponse
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
		return &AdminTokenResponse{}, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := c.client.Do(request)
	if err != nil {
		return &AdminTokenResponse{}, err
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
		return &AdminTokenResponse{}, err
	}

	err = c.rdb.Set(
		context.Background(),
		ADMIN_REFRESH_TOKEN,
		result.RefreshToken,
		time.Duration(result.RefreshTokenExpiredIn-1)*time.Second,
	).Err()
	if err != nil {
		return &AdminTokenResponse{}, err
	}

	return &result, nil
}

type AdminTokenResponse struct {
	AccessToken           string `json:"access_token"`
	ExpiredIn             int    `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiredIn int    `json:"refresh_expires_in"`
	SessionState          string `json:"session_state"`
	Scope                 string `json:"scope"`
}
