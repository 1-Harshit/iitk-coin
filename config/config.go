package config

import (
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"log"
)

var Out = log.Println // alias cuz why not :-/
var JwtKey = []byte("a5=?4K59Wnk=k#cYwG@ZZwsM56rVFDew")

type User struct {
	Roll     int    `json:"roll" validate:"required" sql:"roll"`
	Name     string `json:"name" validate:"required" sql:"name"`
	Email    string `json:"email" sql:"email"`
	Password string `json:"password" validate:"required" sql:"password"`
}

type Wallet struct {
	Roll  int `json:"roll" sql:"roll"`
	Coins int `json:"coins" sql:"coins"`
}

type Trnxn struct {
	From  int `json:"from" sql:"from"`
	To    int `json:"to" sql:"to"`
	Coins int `json:"coins" sql:"coins"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func HashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	Check(err)
	return string(hash)
}

func JWTKey(token *jwt.Token) (interface{}, error) {
	return JwtKey, nil
}

func Check (e error)  {
	if e != nil {
		panic(e)
	}
}