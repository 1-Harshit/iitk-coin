package config

import (
	"strings"
)

// Check non empty request and email
func ValidateUserforOTP(t User) string {
	if t.Roll <= 0 {
		return "Roll Number not found in request"
	}
	if t.Name == "" {
		return "Name not found in request"
	}
	if t.Email == "" {
		return "Email not found in request"
	}
	if !strings.HasSuffix(t.Email, "@iitk.ac.in") {
		return "IITK Email ID not found in request"
	}
	// return checkEmailandRoll(t.Email, t.Roll)
	return ""
}

// Check non empty request and email
func ValidateUser(t User) string {
	if t.Roll <= 0 {
		return "Roll Number not found in request"
	}
	if t.Name == "" {
		return "Name not found in request"
	}
	if t.Email == "" {
		return "Email not found in request"
	}
	if !strings.HasSuffix(t.Email, "@iitk.ac.in") {
		return "IITK Email ID not found in request"
	}
	if t.Password == "" {
		return "Password not found in request"
	}
	if t.Batch == 0 {
		return "Batch not found in request"
	}
	if t.OTP == "" {
		return "OTP not found in the request"
	}
	return ""
}

func ValidateCredentials(t User) string {
	if t.Roll == 0 {
		return "Roll Number not found in request"
	}
	if t.Password == "" {
		return "Password Number not found in request."
	}
	return ""
}

func ValidateReward(t Wallet) string{
	if t.Roll == 0 {
		return "Roll Number not found in request"
	}
	if t.Coins < 0 {
		return "Positive coins needed in request"
	}
	if x := 100*t.Coins; x == float64(int(x)){
		return "Coins only allowed till two decimal places"
	}
	return ""
}

func ValidateItem(t Item) string{
	if t.Name == "" {
		return "Item Name not found in request"
	}
	if t.Value < 0 {
		return "Positive Value of item needed in request"
	}
	return ""
}

func CalculateTax(t Wallet, f *Claims) float64 {
	var tax float64
	if t.Batch == f.Batch {
		tax = IntraBatchTax*t.Coins
	}else {
		tax = InterBatchTax*t.Coins
	}
	return tax
}