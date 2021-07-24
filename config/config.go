package config

import (
	"github.com/golang-jwt/jwt"
)

var JwtKey = []byte("xJ+Ln7CGGswS?EUsq+GJ*6@q%WN9mtLTbN*5kVd2BtTsN8#Gf@9c")

type User struct {
	Roll     int    `json:"roll" validate:"required" sql:"roll"`
	Name     string `json:"name" validate:"required" sql:"name"`
	Email    string `json:"email" sql:"email"`
	Password string `json:"password" validate:"required" sql:"password"`
	Batch    int    `json:"batch" sql:"batch"`
	UsrType  int    `json:"usrtype" sql:"usrtype"`
	OTP      string `json:"otp" sql:"otp"`
}

type Wallet struct {
	Roll    int     `json:"roll" sql:"roll"`
	Coins   float64 `json:"coins" sql:"coins"`
	UsrType int     `json:"usrtype" sql:"usrtype"`
	Batch   int     `json:"batch" sql:"batch"`
	OTP     string  `json:"otp"`
	Rem     string  `json:"remarks" sql:"remarks"`
}

type Trnxn struct {
	From  int     `json:"from" sql:"from"`
	To    int     `json:"to" sql:"to"`
	Coins float64 `json:"coins" sql:"coins"`
	Rem   string  `json:"remarks" sql:"remarks"`
}

type Claims struct {
	Roll    int `json:"roll"`
	UsrType int `json:"usrtype"`
	Batch   int `json:"batch"`
	jwt.StandardClaims
}

type Response struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type RespToken struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	JwtToken string `json:"jwttoken"`
}

type RespData struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type Item struct {
	ItemNo int    `json:"itemNo"`
	Name   string `json:"name"`
	Value  int    `json:"value"`
}

type Redeem struct {
	Id     int    `json:"redeemid"`
	Roll   int    `json:"roll"`
	ItemNo int    `json:"itemNo"`
	Status int    `json:"status"`
	Time   string `json:"time"`
	Name   string `json:"name"`
	Value  int    `json:"value"`
	OTP    string `json:"otp"`
}

type Reward struct {
	Time	string `json:"time"`
	Coins	int    `json:"coins"`
	Remarks string `json:"remarks"`
}

type Tnxn struct {
	Time	string `json:"time"`
	From	int    `json:"from"`
	To		int    `json:"to"`
	Sent 	int    `json:"sent"`
	Tax	    float64`json:"tax"`
	Remarks string `json:"remarks"`
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}
