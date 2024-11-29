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

	http.HandleFunc("/admin/token", handler.GetAdminTokenHandler)

	log.Println("server started at :2901...")
	log.Fatal(http.ListenAndServe(":2901", nil))
}
