package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var dbpath = "./data.db"
var out = log.Println // alias cuz why not :-/
var jwtKey = []byte("a5=?4K59Wnk=k#cYwG@ZZwsM56rVFDew")
var Mydb *sql.DB

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

func main() {

	out("let's Go!")
	var err error
	Mydb, err = sql.Open("sqlite3", dbpath)
	check(err)

	// make the table 'User' and 'Wallet'
	createTable()

	mux := http.NewServeMux()

	mux.HandleFunc("/", hello)
	mux.HandleFunc("/signup", signUp)
	mux.HandleFunc("/login", Login)
	mux.HandleFunc("/secretpage", SecretPage)
	mux.HandleFunc("/reward", Reward)
	mux.HandleFunc("/transfer", Transfer)
	mux.HandleFunc("/view", View)

	out("Starting server. Listening on port http://localhost:8080")

	err = http.ListenAndServe(":8080", mux)
	check(err)
}

func createTable() {

	command := `CREATE TABLE IF NOT EXISTS "User" (
		"roll"	INTEGER NOT NULL,
		"name"	TEXT NOT NULL,
		"email"	TEXT NOT NULL UNIQUE,
		"password"	TEXT NOT NULL,
		"createdat"	TEXT,
		PRIMARY KEY("roll")
	);`
	// check the command
	statement, err := Mydb.Prepare(command)
	check(err)

	statement.Exec()

	command = `CREATE TABLE IF NOT EXISTS "Wallet" (
		"sl"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"roll"	INTEGER NOT NULL UNIQUE,
		"coins"	INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY("roll") REFERENCES "User"("roll")
	);`

	statement, err = Mydb.Prepare(command)
	check(err)

	statement.Exec()
	out("Tables created")
}

func insertIT(dt User) error {

	insert_user := `INSERT INTO "main"."User"
		("roll", "name", "email", "password", "createdat")
		VALUES (?, ?, ?, ?, ?);`

	statement, err := Mydb.Prepare(insert_user)
	if err != nil {
		return err
	}

	_, err = statement.Exec(dt.Roll, dt.Name, dt.Email, dt.Password, time.Now().Format(time.RFC1123Z))
	if err != nil {
		return err
	}

	insert_wal := `INSERT INTO "main"."Wallet"
		("roll")
		VALUES (?);`

	statement, err = Mydb.Prepare(insert_wal)
	if err != nil {
		return err
	}

	_, err = statement.Exec(dt.Roll)
	if err != nil {
		return err
	}

	return nil
}

func GetUser(roll int) (User, error) {
	var usr User

	err := Mydb.QueryRow(`SELECT "roll", "name", "email", "password" FROM "main"."User" WHERE roll = $1`, roll).Scan(&usr.Roll, &usr.Name, &usr.Email, &usr.Password)
	if err != nil {
		return User{}, err
	} else {
		return usr, nil
	}
}

func GetCoins(roll int) (int, error) {

	var coins int
	err := Mydb.QueryRow(`SELECT "coins" FROM "main"."Wallet" WHERE roll = $1`, roll).Scan(&coins)
	if err != nil {
		return -1, err
	} else {
		return coins, nil
	}
}

func RewardCoins(x Wallet) error {

	upd := `UPDATE "main"."Wallet" 
		SET coins= coins + ? 
		WHERE "roll"=?;;`

	statement, err := Mydb.Prepare(upd)
	if err != nil {
		return err
	}

	stmt, err := statement.Exec(x.Coins, x.Roll)
	if err != nil {
		return err
	}
	count, err2 := stmt.RowsAffected() 
	if err2 != nil {
		return err2
	}
	if count == 0 {
		return errors.New("no Such roll found")
	}
	return nil
}

func TransferCoins(t Trnxn) error {

	if _, err := GetCoins(t.From); err != nil{
		return errors.New("invalid sender roll")
	}
	if _, err := GetCoins(t.To); err != nil{
		return errors.New("invalid reciever roll")
	}

	tx, err := Mydb.Begin()
	if err != nil {
		return err
	}
	
	upd := `UPDATE "main"."Wallet" 
		SET coins= coins - $1 
		WHERE "roll"=$2 AND coins >$1;;`

	statement, err := tx.Prepare(upd)
	if err != nil {
		tx.Rollback()
		return err
	}

	stmt, err := statement.Exec(t.Coins, t.From)
	if err != nil {
		tx.Rollback()
		return err
	}
	count, err2 := stmt.RowsAffected() 
	if err2 != nil {
		tx.Rollback()
		return err2
	}
	if count == 0{
		tx.Rollback()
		return errors.New("sender wallet doesn't have that capacity")
	}
	_, err = statement.Exec(-t.Coins, t.To)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()

	return err
}

func hello(res http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		fmt.Fprintln(res, "Hello World!")
	}
}

func signUp(rw http.ResponseWriter, req *http.Request) {

	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {
		dec := json.NewDecoder(req.Body)
		var t User
		err := dec.Decode(&t)

		if isNotValid := validateUser(err, rw, t); isNotValid {
			return
		}

		t.Password = hashAndSalt([]byte(t.Password))

		err = insertIT(t)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Error in signing up.\n")
			return
		}
		rw.WriteHeader(http.StatusOK)
		fmt.Fprintf(rw, "User %s with roll %d is created\n", t.Name, t.Roll)
	}
}

func Login(rw http.ResponseWriter, req *http.Request) {

	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {
		dec := json.NewDecoder(req.Body)
		var t User
		err := dec.Decode(&t)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Error in decoding data\n")
			return
		}

		isNotValidCred := validateCredentials(t, rw)
		if isNotValidCred {
			return
		}

		dt, err := GetUser(t.Roll)
		if err != nil {
			if err == sql.ErrNoRows {
				rw.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(rw, "No such user found\n")
				return
			}
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			fmt.Fprintf(rw, "Internal sever error bruh\n")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(dt.Password), []byte(t.Password))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			fmt.Fprintf(rw, "Wrong Password\n")
			return
		}
		fmt.Fprintf(rw, "Hey, User %s! Your roll is %d.\n", dt.Name, dt.Roll)

		expirationTime := time.Now().Add(5 * time.Minute)

		claims := &Claims{
			Username: dt.Name,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(rw, "Your JWT token is: %s\n", tokenString)
		fmt.Fprintf(rw, "Valid for next 5 Mins\n")
	}
}

func SecretPage(rw http.ResponseWriter, r *http.Request) {

	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		http.Error(rw, "No Token found", http.StatusBadRequest)
		return
	}
	tknStr := strings.TrimSpace(splitToken[1])

	claims := &Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, JWTKey)
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			http.Error(rw, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	if !tkn.Valid {
		http.Error(rw, err.Error(), http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(rw, "Welcome %s!\n", claims.Username)
	fmt.Fprintf(rw, "You have successfully accessed this secretpage!\n")
}

func Reward(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {
		dec := json.NewDecoder(req.Body)
		var t Wallet
		err := dec.Decode(&t)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Error in decoding data\n")
			return
		}

		if t.Roll == 0 {
			http.Error(rw, "No Roll found", http.StatusBadRequest)
			fmt.Fprintf(rw, "Roll Number not found in request\n")
			return
		}

		err = RewardCoins(t)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "annont reward coins\n")
			return
		}
		fmt.Fprint(rw, "Successfully Rewarded\n")
	}
}

func Transfer(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {
		dec := json.NewDecoder(req.Body)
		var t Trnxn
		err := dec.Decode(&t)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Error in decoding data\n")
			return
		}

		isInvalidRequest := preValidateTrnxn(t, rw)
		if isInvalidRequest {
			return
		}

		err = TransferCoins(t)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Couldnot Transfer coins\n")
			return
		}

		fmt.Fprintf(rw, "Successfuly Transferred\n")
	}
}

func View(rw http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {
		dec := json.NewDecoder(req.Body)
		var t Wallet
		err := dec.Decode(&t)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Error in decoding data\n")
			return
		}

		if t.Roll == 0 {
			http.Error(rw, "No Roll found", http.StatusBadRequest)
			fmt.Fprintf(rw, "Roll Number not found in request\n")
			return
		}

		t.Coins, err = GetCoins(t.Roll)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Error in Fetching Coins\n")
			return
		}

		fmt.Fprintf(rw, "Roll %d has %d coin(s)\n", t.Roll, t.Coins)
	}
}

func validateUser(err error, rw http.ResponseWriter, t User) bool {
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

func validateCredentials(t User, rw http.ResponseWriter) bool {
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

func preValidateTrnxn(t Trnxn, rw http.ResponseWriter) bool {
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

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	check(err)
	return string(hash)
}

func JWTKey(token *jwt.Token) (interface{}, error) {
	return jwtKey, nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}