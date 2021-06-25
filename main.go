package main

import (
	"log"
	"net/http"

	r "github.com/1-Harshit/iitk-coin/routes"
)

func main() {
	// server
	mux := http.NewServeMux()

	// all endpoints
	mux.HandleFunc("/", r.Hello)
	mux.HandleFunc("/signup", r.SignUp)
	mux.HandleFunc("/login", r.Login)
	mux.HandleFunc("/secretpage", r.SecretPage)
	mux.HandleFunc("/reward", r.Reward)
	mux.HandleFunc("/transfer", r.Transfer)
	mux.HandleFunc("/view", r.View)

	log.Println("Starting server. Listening on http://localhost:8080")

	// port 8080
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}