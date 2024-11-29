DSN="keycloak:keycloak@tcp(localhost:3306)/keycloak?parseTime=true"

g-up:
	goose -dir ./migrations mysql $(DSN) up

g-down:
	goose -dir ./migrations mysql $(DSN) down