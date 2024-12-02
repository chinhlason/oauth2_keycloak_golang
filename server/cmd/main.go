package main

import (
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	server "oauth2"
	"time"
)

var (
	KEYCLOAK_HOST = "http://localhost:8080"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	keycloakClient := server.NewKeyCloakClient(KEYCLOAK_HOST, 10*time.Second, rdb)
	handler := server.NewHandler(keycloakClient)
	errChan := make(chan error)

	go func() {
		errChan <- keycloakClient.CreateRealms([]string{"users"})
	}()

	if err := <-errChan; err != nil {
		log.Println(err)
	}

	middleware := server.NewMiddleware(keycloakClient)

	http.HandleFunc("/admin/token", handler.GetAdminTokenHandler)
	http.HandleFunc("/realms", handler.CreateRealms)
	http.HandleFunc("/create", handler.CreateUser)
	http.HandleFunc("/set-password", handler.SetPassword)
	http.HandleFunc("/user/token", handler.GetUserToken)
	http.HandleFunc("/introspect", handler.IntrospectToken)
	http.HandleFunc("/refresh", handler.RefreshToken)
	http.Handle("/update", middleware.ValidateToken(http.HandlerFunc(handler.UpdateUserInfo)))

	log.Println("server started at :2901...")
	log.Fatal(http.ListenAndServe(":2901", nil))
}
