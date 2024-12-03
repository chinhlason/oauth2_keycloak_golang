package server

import (
	"fmt"
	"time"
)

const (
	DEFAULT_VALUE = "default"
	DEFAULT_ROLE  = "user"
)

type IService interface {
	LoginWithGoogle(token string) (*TokenResponse, error)
	Register(username, email, password, firstName, lastName string) error
	Login(username, password string) (*TokenResponse, error)
	Logout(refreshToken string) error
}

type Service struct {
	keyCloakClient IKeyCloak
	repository     IRepository
}

func NewService(keyCloakClient IKeyCloak, repository IRepository) IService {
	return &Service{keyCloakClient, repository}
}

func (s *Service) LoginWithGoogle(token string) (*TokenResponse, error) {
	userInfo, err := GetUserInfo(token)
	if err != nil {
		return nil, err
	}

	userFromKeycloak, err := s.keyCloakClient.GetUserByEmail(userInfo["email"].(string))
	if err == nil {
		token, err := s.keyCloakClient.GetUserToken(userFromKeycloak.Username, DEFAULT_VALUE)
		if err != nil {
			return nil, err
		}
		return token, nil
	}

	req := &CreateUserRequest{
		Username:      userInfo["email"].(string),
		Email:         userInfo["email"].(string),
		FirstName:     userInfo["given_name"].(string),
		LastName:      userInfo["family_name"].(string),
		Enabled:       true,
		EmailVerified: false,
	}

	id, err := s.keyCloakClient.CreateUser(DEFAULT_VALUE, req, DEFAULT_VALUE)
	if err != nil {
		return nil, err
	}

	err = s.repository.Save(id,
		userInfo["email"].(string),
		userInfo["email"].(string),
		DEFAULT_VALUE,
		fmt.Sprintf("%s %s", userInfo["given_name"].(string), userInfo["family_name"].(string)),
		DEFAULT_ROLE,
		time.Now(), time.Now())
	if err != nil {
		return nil, err
	}

	tokenResp, err := s.keyCloakClient.GetUserToken(userInfo["email"].(string), DEFAULT_VALUE)
	if err != nil {
		return nil, err
	}
	return tokenResp, nil
}

func (s *Service) Register(username, email, password, firstName, lastName string) error {
	req := &CreateUserRequest{
		Username:      username,
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		Enabled:       true,
		EmailVerified: false,
	}

	id, err := s.keyCloakClient.CreateUser(DEFAULT_VALUE, req, DEFAULT_VALUE)
	if err != nil {
		return err
	}

	err = s.repository.Save(id, username, email, password, fmt.Sprintf("%s %s", firstName, lastName), DEFAULT_ROLE, time.Now(), time.Now())
	if err != nil {
		//Rollback keycloak
		err2 := Retry(3, 500*time.Millisecond, func() error {
			return s.keyCloakClient.DeleteUser(id)
		})
		if err2 != nil {
			return err2
		}
		return err
	}
	return nil
}

func (s *Service) Login(username, password string) (*TokenResponse, error) {
	_, err := s.repository.Get(username)
	if err != nil {
		return nil, err
	}

	token, err := s.keyCloakClient.GetUserToken(username, password)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *Service) Logout(refreshToken string) error {
	return s.keyCloakClient.Logout(refreshToken)
}
