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

// Endpoint to signup
// POST: Roll-int, Name-string,	Email-string, Password-string, Batch-int
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
					Message: "User Not registedfgdred",
				}
				json.NewEncoder(rw).Encode(response)
				// rw.Write(Rsp(err.Error(), "User Not registered"))
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

		// Get the walllet
		wal, err := db.GetWallet(dt.Roll)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			response := c.Response{
				Error:   err.Error(),
				Message: "User Not found",
			}
			json.NewEncoder(rw).Encode(response)
			// rw.Write(Rsp(err.Error(), "User Not found"))
			return
		}
		dt.Batch = wal.Batch
		// get token to
		tokenString, err := auth.GetJwtToken(dt, wal.UsrType)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			response := c.Response{
				Error:   err.Error(),
				Message: "Error While getting JWT token",
			}
			json.NewEncoder(rw).Encode(response)
			// rw.Write(Rsp(err.Error(), "Error While getting JWT token"))
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
		}

		// a tx to pass
		tx := c.Trnxn{
			From:  claims.Roll,
			To:    t.Roll,
			Coins: t.Coins,
			Rem:   t.Rem,
		}

		// add to db
		err = db.RewardCoins(tx)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			a := fmt.Sprintf("Cannont reward %f coins to %d", t.Coins, t.Roll)
			rw.Write(Rsp(err.Error(), a))
			return
		}

		// Sucessful
		rw.WriteHeader(http.StatusOK)
		a := fmt.Sprintf("Successfuly Transfered %f coins to %d", t.Coins, t.Roll)
		rw.Write(Rsp("", a))
	}
}

// Endpoint accessible by all to transfer coins
// POST: roll-int, coins-int & Authentication Bearer Header
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

// Endpoint to get coins
// GET: Authentication Bearer Header
func View(rw http.ResponseWriter, r *http.Request) {
	// get the token
	reqToken := r.Header.Get("Authorization")

	// get the claims
	claims, isNotValid := GetClaims(reqToken, rw)
	if isNotValid {
		return
	}

	// get coins
	Coins, err := db.GetCoins(claims.Roll)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		fmt.Fprintf(rw, "Error in Fetching Coins\n")
		return
	}

	// We're okay
	msg := fmt.Sprintf("Roll %d has %f coins", claims.Roll, Coins)
	rw.WriteHeader(http.StatusOK)
	rw.Write(Rsp("", msg))
}

/*STORE*/

// Endpoint to View all Items
// GET aiwehi
func ListItems(rw http.ResponseWriter, req *http.Request) {
	if c.IsStoreOpen {
		data, err := db.GetItems()
		if err != nil {
			response := c.Response{
				Error:   err.Error(),
				Message: "Can't fetch data",
			}
			rw.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(rw).Encode(response)
		}
		response := c.RespData{
			Message: "All active items are listed",
			Data:    data,
		}
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
	}else{
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(Rsp("Store is closed", "Please come back at a later date"))
	}
}

// AH-GenSec Only
// Endpoint to insert Items
// POST: name-string, value-int & Authentication Bearer Header
func AddItems(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.Item
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
		if valid := c.ValidateItem(t); valid != "" {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(valid, "Bad Request"))
		}

		// add to db
		err = db.InsertItems(t)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Cannot insert Item to store"))
			return
		}

		// Sucessful
		rw.WriteHeader(http.StatusOK)
		rw.Write(Rsp("", "Item Successfully Added"))
	}
}

// AH-GenSec Only
// Endpoint to Delete Items
// POST: itemNo-int & Authentication Bearer Header
func RemoveItems(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.Item
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
		if t.ItemNo == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp("Item No not found in request", "Bad Request"))
		}

		// add to db
		err = db.DeleteItems(t)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Cannot Delete Item from store"))
			return
		}

		// Sucessful
		rw.WriteHeader(http.StatusOK)
		rw.Write(Rsp("", "Item Successfully Deleted"))
	}
}

/*Redeem*/

// Endpoint to Request a Redeem
// POST: itemNo-int & Authentication Bearer Header
func RedeemRequest(rw http.ResponseWriter, req *http.Request){
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.Redeem
		err := dec.Decode(&t)

		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Error in decoding data"))
			return
		}

		// get the token
		reqToken := req.Header.Get("Authorization")
		usr, isNotValid := GetClaims(reqToken, rw)
		if isNotValid {
			return
		}

		// input verification
		if t.ItemNo == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp("ItemNo not found in request", "Bad Request"))
		}

		t.Roll = usr.Roll

		// alter the database
		err = db.ReqRedeem(t)

		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Couldnot Transfer coins"))
			return
		}

		// Reedeem done
		rw.WriteHeader(http.StatusOK)
		rw.Write(Rsp("", "Request sent Successful"))
	}
}

// AH-GenSec Only
// Endpoint to View Req
// GET: Authentication Bearer Header
func ListRedeemRequest(rw http.ResponseWriter, req *http.Request){
	
	// get the token
	reqToken := req.Header.Get("Authorization")
	usr, isNotValid := GetClaims(reqToken, rw)
	if isNotValid {
		return
	}

	if usr.UsrType != 1 {
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(Rsp("Only GenSec and AH can access this endpoint", ""))
		return
	}

	data, err := db.GetReedem()
	if err != nil {
		response := c.Response{
			Error:   err.Error(),
			Message: "Can't fetch data",
		}
		rw.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(rw).Encode(response)
	}
	response := c.RespData{
		Message: "All pending requests are listed",
		Data:    data,
	}
	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)	
}

// AH-GenSec Only
// Endpoint to reject request
// post: id-int & Authentication Bearer Header
func RejectRedeemRequest(rw http.ResponseWriter, req *http.Request){
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.Redeem
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
		if t.Id == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp("ID not found in request", "Bad Request"))
		}

		// Alter DB
		err = db.RejectRedeem(t.Id)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Could not Reject Request"))
			return
		}

		// Sucessful
		rw.WriteHeader(http.StatusOK)
		rw.Write(Rsp("", "Successfuly Rejected"))
	}
}

// AH-GenSec Only
// Endpoint to Accept request
// post: id-int & Authentication Bearer Header
func ApproveRedeemRequest(rw http.ResponseWriter, req *http.Request){
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {

		// Input from request body
		dec := json.NewDecoder(req.Body)
		var t c.Redeem
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
		if t.Id == 0 {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp("ID not found in request", "Bad Request"))
		}

		// Alter DB
		err = db.ApproveRedeem(t.Id)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(Rsp(err.Error(), "Could not Accept Request"))
			return
		}

		// Sucessful
		rw.WriteHeader(http.StatusOK)
		rw.Write(Rsp("", "Successfuly Accepted"))
	}
}