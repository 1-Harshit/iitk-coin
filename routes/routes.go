package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	c "github.com/1-Harshit/iitk-coin/config"
	"github.com/1-Harshit/iitk-coin/db"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

func Hello(res http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		fmt.Fprintln(res, "Hello World!")
	}
}

func SignUp(rw http.ResponseWriter, req *http.Request) {

	if req.Method == "GET" {
		rw.Write([]byte("Only POST request allowed"))
	} else {
		dec := json.NewDecoder(req.Body)
		var t c.User
		err := dec.Decode(&t)

		if isNotValid := c.ValidateUser(err, rw, t); isNotValid {
			return
		}

		t.Password = c.HashAndSalt([]byte(t.Password))

		err = db.InsertIT(t)
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
		var t c.User
		err := dec.Decode(&t)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Error in decoding data\n")
			return
		}

		isNotValidCred := c.ValidateCredentials(t, rw)
		if isNotValidCred {
			return
		}

		dt, err := db.GetUser(t.Roll)
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

		claims := &c.Claims{
			Username: dt.Name,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		tokenString, err := token.SignedString(c.JwtKey)
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

	claims := &c.Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, c.JWTKey)
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
		var t c.Wallet
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

		err = db.RewardCoins(t)
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
		var t c.Trnxn
		err := dec.Decode(&t)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Error in decoding data\n")
			return
		}

		isInvalidRequest := c.PreValidateTrnxn(t, rw)
		if isInvalidRequest {
			return
		}

		err = db.TransferCoins(t)

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
		var t c.Wallet
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

		t.Coins, err = db.GetCoins(t.Roll)

		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)
			fmt.Fprintf(rw, "Error in Fetching Coins\n")
			return
		}

		fmt.Fprintf(rw, "Roll %d has %d coin(s)\n", t.Roll, t.Coins)
	}
}