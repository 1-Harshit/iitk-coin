package routes

import (
	"encoding/json"
	"net/http"

	c "github.com/1-Harshit/iitk-coin/config"
	"github.com/1-Harshit/iitk-coin/db"
	"github.com/1-Harshit/iitk-coin/auth"
)

/*STORE*/

// Endpoint to View all Items
// GET aiwehi
func ListItems(rw http.ResponseWriter, _ *http.Request) {
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
	} else {
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


// Endpoint accessible by all to get otp for Redeeming coins
// POST: Authentication Bearer Header
func OTPforRedeem(rw http.ResponseWriter, req *http.Request) {

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
	if valid := c.Email(t, OTP, 3); valid != "" {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(Rsp(valid, "Error in sending Email"))
	}

	
	// Everything went well
	rw.WriteHeader(http.StatusOK)
	rw.Write(Rsp("", "email Sent"))
}

// Endpoint to Request a Redeem
// POST: itemNo-int & Authentication Bearer Header
func RedeemRequest(rw http.ResponseWriter, req *http.Request) {
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
func ListRedeemRequest(rw http.ResponseWriter, req *http.Request) {

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
func RejectRedeemRequest(rw http.ResponseWriter, req *http.Request) {
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
func ApproveRedeemRequest(rw http.ResponseWriter, req *http.Request) {
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
