package main

import (
	"log"
	"net/http"

	r "github.com/1-Harshit/iitk-coin/routes"
)

var out = log.Println // alias cuz why not :-/


func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", r.Hello)
	mux.HandleFunc("/signup", r.SignUp)
	mux.HandleFunc("/login", r.Login)
	mux.HandleFunc("/secretpage", r.SecretPage)
	mux.HandleFunc("/reward", r.Reward)
	mux.HandleFunc("/transfer", r.Transfer)
	mux.HandleFunc("/view", r.View)

	out("Starting server. Listening on port http://localhost:8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}