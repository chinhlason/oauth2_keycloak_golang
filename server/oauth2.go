package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"os"

	"net/http"
)

var googleOauthConfig = &oauth2.Config{}

func init() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	googleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_SECRET_CODE"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URIS"),
		Scopes:       []string{"profile", "email"},
		Endpoint:     google.Endpoint,
	}
}

func GetTokenFromCode(code string) (*oauth2.Token, error) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func HandleCallback(res http.ResponseWriter, req *http.Request) {
	code := req.URL.Query().Get("code")
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	userInfo, err := GetUserInfo(token.AccessToken)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	responseJson, _ := json.Marshal(userInfo)
	res.WriteHeader(http.StatusOK)
	res.Write(responseJson)
}

func GetUserInfo(accessToken string) (map[string]interface{}, error) {
	userInfoEndpoint := "https://www.googleapis.com/oauth2/v2/userinfo"
	resp, err := http.Get(fmt.Sprintf("%s?access_token=%s", userInfoEndpoint, accessToken))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}
	return userInfo, nil
}
