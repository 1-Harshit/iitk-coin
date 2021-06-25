package config

import (
	"github.com/dgrijalva/jwt-go"
)

var JwtKey = []byte("xJ+Ln7CGGswS?EUsq+GJ*6@q%WN9mtLTbN*5kVd2BtTsN8#Gf@9c")

type User struct {
	Roll	 int    `json:"roll" validate:"required" sql:"roll"`
	Name	 string `json:"name" validate:"required" sql:"name"`
	Email	 string `json:"email" sql:"email"`
	Password string `json:"password" validate:"required" sql:"password"`
	Batch	 int 	`json:"batch" sql:"batch"`
}

type Wallet struct {
	Roll  	int 	`json:"roll" sql:"roll"`
	Coins 	float64 `json:"coins" sql:"coins"`
	UsrType int 	`json:"usrtype" sql:"usrtype"`
	Batch 	int 	`json:"batch" sql:"batch"`
	Rem		string  `json:"remarks" sql:"remarks"`
}

type Trnxn struct {
	From  int 		`json:"from" sql:"from"`
	To    int 		`json:"to" sql:"to"`
	Coins float64 	`json:"coins" sql:"coins"`
	Rem		string  `json:"remarks" sql:"remarks"`
}

type Claims struct {
	Roll		int 	`json:"roll"`
	UsrType 	int 	`json:"usrtype"`
	Batch		int		`json:"batch"`
	jwt.StandardClaims
}

type Response struct{
	Error	string	`json:"error"`
	Message	string	`json:"message"`
}

type RespToken struct{
	Error	string	`json:"error"`
	Message	string	`json:"message"`
	JwtToken string	`json:"jwttoken"`
}

func Check (e error)  {
	if e != nil {
		panic(e)
	}
}

