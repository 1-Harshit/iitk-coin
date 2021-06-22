package config

import (
	"fmt"
	"net/http"
	"strings"
)
func ValidateUser(err error, rw http.ResponseWriter, t User) bool {
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		fmt.Fprintf(rw, "Error in decoding data\n")
		return true
	}
	if t.Roll == 0 {
		http.Error(rw, "No Roll found", http.StatusBadRequest)
		fmt.Fprintf(rw, "Roll Number not found in request\n")
		return true
	}
	if t.Name == "" {
		http.Error(rw, "No Name found", http.StatusBadRequest)
		fmt.Fprintf(rw, "Name not found in request\n")
		return true
	}
	if t.Email == "" {
		http.Error(rw, "No Email found", http.StatusBadRequest)
		fmt.Fprintf(rw, "Email not found in request\n")
		return true
	}
	if !strings.HasSuffix(t.Email, "@iitk.ac.in") {
		http.Error(rw, "IITK Email Required", http.StatusBadRequest)
		fmt.Fprintf(rw, "Wrong Email found in request\n")
		return true
	}
	if t.Password == "" {
		http.Error(rw, "Password cannot be empty", http.StatusBadRequest)
		fmt.Fprintf(rw, "Password not found\n")
		return true
	}
	return false
}

func ValidateCredentials(t User, rw http.ResponseWriter) bool {
	if t.Roll == 0 {
		http.Error(rw, "No Roll found", http.StatusBadRequest)
		fmt.Fprintf(rw, "Roll Number not found in request\n")
		return true
	}
	if t.Password == "" {
		http.Error(rw, "No Password found", http.StatusBadRequest)
		fmt.Fprintf(rw, "Password Number not found\n")
		return true
	}
	return false
}

func PreValidateTrnxn(t Trnxn, rw http.ResponseWriter) bool {
	if t.From == 0 {
		http.Error(rw, "No Sender found", http.StatusBadRequest)
		fmt.Fprintf(rw, "Sender Roll Number not found in request\n")
		return true
	}
	if t.To == 0 {
		http.Error(rw, "No Reciever found", http.StatusBadRequest)
		fmt.Fprintf(rw, "Reciever Roll Number not found in request\n")
		return true
	}
	if t.Coins < 0 {
		http.Error(rw, "Coins should be positive", http.StatusBadRequest)
		fmt.Fprintf(rw, "positive coins needed in request\n")
		return true
	}
	return false
}