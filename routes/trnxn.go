package routes

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/1-Harshit/iitk-coin/auth"
	c "github.com/1-Harshit/iitk-coin/config"
	"github.com/1-Harshit/iitk-coin/db"
)

// Endpoint only accessible by Gensec AH
// POST: Roll-int, Coins-int & Authentication Bearer Header
func Reward(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.Wallet
		err := dec.Decode(&t)

		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Error in decoding data"))
			return
		}

		// get the token
		reqToken := req.Header.Get("Authorization")
		claims, isNotValid := GetClaims(reqToken, rw)
		if isNotValid {
			return
		}

		// check if usr is Gensec or AH
		if claims.UsrType != 1 {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(Rsp("Only GenSec and AH can access this endpoint", ""))
			return
		}

		// input verification
		if valid := c.ValidateReward(t); valid != "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(valid, "Bad Request"))
			return
		}

		// add to db
		err = db.RewardCoins(t)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			a := fmt.Sprintf("Cannont reward %f coins to %d", t.Coins, t.Roll)
			rw.Write(Rsp(err.Error(), a))
			return
		}

		// Sucessful
		rw.WriteHeader(http.StatusOK)
		a := fmt.Sprintf("Successfuly Rewarded %f coins to %d", t.Coins, t.Roll)
		rw.Write(Rsp("", a))
	}
}

// Endpoint accessible by all to get otp for transfering coins
// POST: Authentication Bearer Header
func OTPforTransfer(rw http.ResponseWriter, req *http.Request) {

	// get the token
	reqToken := req.Header.Get("Authorization")
	usr, isNotValid := GetClaims(reqToken, rw)
	if isNotValid {
		return
	}
	// Get existing data on user
	t, err := db.GetUser(usr.Roll)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(Rsp(err.Error(), "Server Error"))
		return
	}
	

	// Check Spam
	if db.ExceedMaxOTP(t.Roll) {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(Rsp("too frequent otp requests", "please try again after 5 minutes"))
	}

	// Get OTP
	OTP := c.GenerateOTP()

	// Hash and salt OTP
	t.OTP = auth.HashAndSalt([]byte(OTP))

	// Insert in DB
	if valid := db.StoreOTP(t); valid != "" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(Rsp(valid, "Error in Storing OTP"))
		return
	}

	// emailing OTP
	if valid := c.Email(t, OTP, 2); valid != "" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(Rsp(valid, "Error in sending Email"))
	}

	
	// Everything went well
	rw.WriteHeader(http.StatusOK)
	rw.Write(Rsp("", "email Sent"))
}

// Endpoint accessible by all to transfer coins
// POST: roll-int, coins-int, otp-string & Authentication Bearer Header
func Transfer(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.Wallet
		err := dec.Decode(&t)

		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Error in decoding data"))
			return
		}

		// get the token
		reqToken := req.Header.Get("Authorization")
		f, isNotValid := GetClaims(reqToken, rw)
		if isNotValid {
			return
		}

		// input verification
		if valid := ValidateTransfer(&t); valid != "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(valid, "Bad Request"))
		}

		// check otp
		hashotp, otpID, err := db.GetOTP(t.Roll)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Error in fetching OTP"))
			return
		}

		// check if otp is valid
		if err := auth.Verify(t.OTP, hashotp); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Error in Verifying OTP"))
			return
		}

		defer db.MarkOTP(otpID)

		// alter the database
		err = db.TransferCoins(t, f)

		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Couldnot Transfer coins"))
			return
		}
		// transaction op
		rw.WriteHeader(http.StatusOK)
		a := fmt.Sprintf("Successfuly Transfered %f coins to %d from %d", t.Coins, t.Roll, f.Roll)
		rw.Write(Rsp("", a))
	}
}
