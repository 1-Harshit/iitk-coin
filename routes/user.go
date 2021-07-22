package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/1-Harshit/iitk-coin/auth"
	c "github.com/1-Harshit/iitk-coin/config"
	"github.com/1-Harshit/iitk-coin/db"
)

// Greeting page on homepage
func Hello(res http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		fmt.Fprintln(res, "Hello World!")
		// } else {
		fmt.Fprintln(res, c.GenerateOTP())
	}
}

// Endpoint to get OTP for signup
// POST: Roll-int, Name-string,	Email-string
func SignUpOTP(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.User
		err := dec.Decode(&t)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Error in Decoding Data"))
			return
		}

		// Validate the request if it is empty
		if valid := c.ValidateUserforOTP(t); valid != "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(valid, "Bad Request"))
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
		if valid := c.Email(t, OTP, -1); valid != "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(valid, "Error in sending Email"))
		}

		
		// Everything went well
		rw.WriteHeader(http.StatusOK)
		rw.Write(Rsp("", "email Sent"))
	}
}

// Endpoint to signup
// POST: Roll-int, Name-string,	Email-string, OTP-string, Password-string, Batch-int
func SignUp(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.User
		err := dec.Decode(&t)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Error in Decoding Data"))
			return
		}

		// Validate the request if it is empty
		if valid := c.ValidateUser(t); valid != "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(valid, "Bad Request"))
			return
		}

		// Hash password
		t.Password = auth.HashAndSalt([]byte(t.Password))

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

		// Insert in DB
		err = db.InsertIT(t)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Error in Signing Up"))
			return
		}

		// Everything went well
		rw.WriteHeader(http.StatusOK)
		rw.Write(Rsp("", fmt.Sprintf("User %s with roll %d is created", t.Name, t.Roll)))
	}
}

// Endpoint to signin
// POST: Roll-int, Password-string
func Login(rw http.ResponseWriter, req *http.Request) {

	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.User
		err := dec.Decode(&t)

		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Error in Decoding Data"))
			return
		}

		// Check if all parametes are there
		if Valid := c.ValidateCredentials(t); Valid != "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Bad Request"))
			return
		}

		// Get existing data on user
		dt, err := db.GetUser(t.Roll)
		if err != nil {
			if err == sql.ErrNoRows {
				rw.WriteHeader(http.StatusUnauthorized)
				response := c.Response{
					Error:   err.Error(),
					Message: "User Not registered",
				}
				json.NewEncoder(rw).Encode(response)
				return
			}
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write(Rsp(err.Error(), "Server Error"))
			return
		}

		// check if password is right
		err = auth.Verify(t.Password, dt.Password)
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write(Rsp(err.Error(), "Wrong Password"))
			return
		}

		// All good yaayy
		rw.WriteHeader(http.StatusOK)
		msg := fmt.Sprintf("Hey, User %s! Your roll is %d. ", dt.Name, dt.Roll)

		// get token to
		tokenString, err := auth.GetJwtToken(dt)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			response := c.Response{
				Error:   err.Error(),
				Message: "Error While getting JWT token",
			}
			json.NewEncoder(rw).Encode(response)
			return
		}
		rw.Write(RspToken("", msg+"This JWT Token Valid for next 5 Minutes", tokenString))
	}
}

// Endpoint to verify Login
// GET: Authentication Bearer Header
func SecretPage(rw http.ResponseWriter, r *http.Request) {
	// get the token
	reqToken := r.Header.Get("Authorization")

	// get the claims
	claims, isNotValid := GetClaims(reqToken, rw)
	if isNotValid {
		return
	}

	// We're okay
	msg := fmt.Sprintf("Welcome %d! You have successfully accessed this secretpage!", claims.Roll)
	rw.WriteHeader(http.StatusOK)
	rw.Write(Rsp("", msg))
}

// Endpoint to get User's Info
// GET: Authentication Bearer Header
func View(rw http.ResponseWriter, r *http.Request) {
	// get the token
	reqToken := r.Header.Get("Authorization")

	// get the claims
	claims, isNotValid := GetClaims(reqToken, rw)
	if isNotValid {
		return
	}

	usr, err := db.GetUser(claims.Roll)

	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(Rsp(err.Error(), "Server Error"))
		return
	}
	rw.WriteHeader(http.StatusOK)
	res := c.RespData{
		Message: fmt.Sprintf("Hey, User %s!", usr.Name),
		Data: usr,
	}
	json.NewEncoder(rw).Encode(res)
}

// Endpoint to get User's Reward
// GET: Authentication Bearer Header
func ViewReward(rw http.ResponseWriter, r *http.Request) {
	// get the token
	reqToken := r.Header.Get("Authorization")
	
	// get the claims
	claims, isNotValid := GetClaims(reqToken, rw)
	if isNotValid {
		return
	}

	dt, err := db.GetUserRewards(claims.Roll)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(Rsp(err.Error(), "Server Error"))
		return
	}
	rw.WriteHeader(http.StatusOK)
	res := c.RespData{
		Message: "All data",
		Data: dt,
	}
	json.NewEncoder(rw).Encode(res)
}

// Endpoint to get User's Transaction
// GET: Authentication Bearer Header
func ViewTransaction(rw http.ResponseWriter, r *http.Request) {
	// get the token
	reqToken := r.Header.Get("Authorization")

	// get the claims
	claims, isNotValid := GetClaims(reqToken, rw)
	if isNotValid {
		return
	}

	dt, err := db.GetUserTransaction(claims.Roll)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(Rsp(err.Error(), "Server Error"))
		return
	}
	rw.WriteHeader(http.StatusOK)
	res := c.RespData{
		Message: "All data",
		Data: dt,
	}
	json.NewEncoder(rw).Encode(res)
}

// Endpoint to get User's Redeem
// GET: Authentication Bearer Header
func ViewRedeem(rw http.ResponseWriter, r *http.Request) {
	// get the token
	reqToken := r.Header.Get("Authorization")

	// get the claims
	claims, isNotValid := GetClaims(reqToken, rw)
	if isNotValid {
		return
	}

	dt, err := db.GetUserRedeem(claims.Roll)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write(Rsp(err.Error(), "Server Error"))
		return
	}
	rw.WriteHeader(http.StatusOK)
	res := c.RespData{
		Message: "All data",
		Data: dt,
	}
	json.NewEncoder(rw).Encode(res)
}