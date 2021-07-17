package config

import (
	"fmt"
	"math/rand"
	"net/smtp"
	"time"
)

/*
// Email username
var From string = "..."

// Email password
var Password string = "..."
*/

// Auth for email
var auth smtp.Auth
// Random source
var src rand.Source

// smtp server configuration.
var smtpHost string = "smtp.gmail.com"
var smtpPort string = "587"

// Charset for otp
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func init() {
	// Authentication.
	auth = smtp.PlainAuth("", From, Password, smtpHost)
	rand.Seed(time.Now().UnixNano())
	src = rand.NewSource(time.Now().UnixNano())
}

func Email(usr User, otp string, reason int) string{
	// why send mail?
	about := ""
	if reason == -1 {
		about = "Use this OTP for Signing Up"
	}else if reason == 1 {
		about = "Use this OTP for changing the password"
	}else if reason == 2 {
		about = "Use this OTP for Transaction"
	}else if reason == 3 {
		about = "Use this OTP for reedeming koins"
	}

	// message
	msg := []byte(
		"To: " + usr.Name + "<" + usr.Email + ">\r\n" +
		"Subject: [Koins] OTP for Koins\r\n" +

		"From: Koins Automation<mailspenpal@gmail.com>\r\n" + "\r\n" +

		"Hi " + usr.Name + " ("+ fmt.Sprintf("%d",usr.Roll) + ")\r\n\r\n" +

		about  + "\r\n\r\n" +

		"This OTP is valid for 5 mins don't share it with anyone:\r\n"+
    	"\tOTP: \t" + otp + "\r\n\r\n"+

		"Please contact Koins admin if you didn't request for this.\r\n\r\n"+

		"Best\r\n"+
		"Koins Automation"+
	"")  

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, From, []string{usr.Email,}, msg)

	if err != nil{
		return err.Error()
	}
	return ""
}

// Generate random string
func GenerateOTP() string {
	n := 10
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}