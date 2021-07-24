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
	
	// Signup
	mux.HandleFunc("/signup", r.SignUp)
	mux.HandleFunc("/signup/otp", r.SignUpOTP)

	// Login
	mux.HandleFunc("/login", r.Login)
	mux.HandleFunc("/secretpage", r.SecretPage)
	
	// forgotpass
	mux.HandleFunc("/forgotpass/otp", r.ForgotPassOTP)
	mux.HandleFunc("/forgotpass", r.ForgotPass)
	
	// Reward AH Gensec
	mux.HandleFunc("/reward", r.Reward)
	
	// Transfer
	mux.HandleFunc("/transfer", r.Transfer)
	mux.HandleFunc("/transfer/otp", r.OTPforTransfer)
	
	// User's info
	mux.HandleFunc("/user/info", r.View)
	mux.HandleFunc("/user/reward", r.ViewReward)
	mux.HandleFunc("/user/transaction", r.ViewTransaction)
	mux.HandleFunc("/user/redeem", r.ViewRedeem)

	// Store
	mux.HandleFunc("/store/list", r.ListItems)
	// Store AH GenSec
	mux.HandleFunc("/store/add", r.AddItems)
	mux.HandleFunc("/store/remove", r.RemoveItems)
	
	
	// Redeem 
	mux.HandleFunc("/redeem/request", r.RedeemRequest)
	mux.HandleFunc("/redeem/request/otp", r.OTPforRedeem)
	// Redeem AH GenSec
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