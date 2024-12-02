package main

import (
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	server "oauth2"
	"time"
)

var (
	KEYCLOAK_HOST = "http://localhost:8080"
	DSN           = "keycloak:keycloak@tcp(localhost:3306)/keycloak?parseTime=true"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	db, err := gorm.Open(mysql.Open(DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("error connecting to database ", err)
	}

	keycloakClient := server.NewKeyCloakClient(KEYCLOAK_HOST, 10*time.Second, rdb)
	repo := server.NewRepository(db)
	service := server.NewService(keycloakClient, repo)
	handler := server.NewHandler(service)
	errChan := make(chan error)

	go func() {
		errChan <- keycloakClient.CreateRealms([]string{"users"})
	}()

	if err := <-errChan; err != nil {
		log.Println(err)
	}

	http.HandleFunc("/callback", handler.LoginWithGoogle)

	//middleware := server.NewMiddleware(keycloakClient)

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(http.DefaultServeMux)

	log.Println("server started at :2901...")
	log.Fatal(http.ListenAndServe(":2901", corsHandler))
}
