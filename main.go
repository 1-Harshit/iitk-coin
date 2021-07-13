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

	// Store
	mux.HandleFunc("/store/list", r.ListItems)
	mux.HandleFunc("/store/add", r.AddItems)
	mux.HandleFunc("/store/remove", r.RemoveItems)
	
	// redeem
	mux.HandleFunc("/redeem/request", r.RedeemRequest)
	mux.HandleFunc("/redeem/list", r.ListRedeemRequest)
	mux.HandleFunc("/redeem/reject", r.RejectRedeemRequest)
	mux.HandleFunc("/redeem/approve", r.ApproveRedeemRequest)

	log.Println("Starting server. Listening on http://localhost:8080")

	// port 8080
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}