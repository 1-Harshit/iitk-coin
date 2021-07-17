package routes

import (
	"encoding/json"
	"net/http"
	"fmt"

	"github.com/1-Harshit/iitk-coin/auth"
	c "github.com/1-Harshit/iitk-coin/config"
	"github.com/1-Harshit/iitk-coin/db"
)


func GetClaims(reqToken string, rw http.ResponseWriter) (*c.Claims, bool) {
	claims := &c.Claims{}

	status, err := auth.AuthenticateToken(reqToken, claims)

	if status == 1 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(Rsp("No Token found", "No Token found"))
		return nil, true
	}

	if status == 2 {
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(Rsp(err.Error(), "User Not Authorised"))
		return nil, true
	}

	if status == 3 {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write(Rsp(err.Error(), "Bad Request"))
		return nil, true
	}

	wal, err := db.GetWallet(claims.Roll)
	if err != nil{
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write(Rsp(err.Error(), "No such user found"))
		return nil, true
	}
	if wal.UsrType != claims.UsrType || wal.Batch != claims.Batch{
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, wal.UsrType != claims.UsrType , wal.Batch != claims.Batch)
		fmt.Fprintln(rw, wal.UsrType , claims.UsrType , wal.Batch , claims.Batch)
		rw.Write(Rsp("Why did u tamper with tokens?", "Please try again"))
		return nil, true
	}
	return claims, false
}

// Write a response in JSON format
func Rsp(err string, message string) []byte {
	re := c.Response{
		Error: err, 
		Message: message,
	}
	resp, err1 := json.Marshal(re)
	c.Check(err1)
	return resp
}

// Write json with a token
func RspToken(err string, message string, jwt string) []byte {
	re := c.RespToken{
		Error: err,
		Message: message,
		JwtToken: jwt,
	}
	resp, err1 := json.Marshal(re)
	c.Check(err1)
	return resp
}

func ValidateTransfer(x *c.Wallet) string {
	if x.Roll <= 0 {
		return "Roll Number is not valid"
	}
	wal, err := db.GetWallet(x.Roll)
	if err != nil{
		return err.Error() + "No such user found"
	}
	if x.Coins <= 0 {
		return "Positive coins needed in request"
	}
	if p :=x.Coins; p != float64(int(p)) {
		return "Integer Coins needed"
	}
	x.UsrType = wal.UsrType
	x.Batch = wal.Batch
	return ""
}