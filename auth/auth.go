package auth

import (
	c "github.com/1-Harshit/iitk-coin/config"
	"time"
	"strings"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

func GetJwtToken(usr c.User) (string, error) {
	expirationTime := time.Now().Add(12 * time.Hour)

	claims := &c.Claims{
		Roll: usr.Roll,
		Batch: usr.Batch,
		UsrType: usr.UsrType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt: time.Now().Unix(),
			Issuer: "Auth.go",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(c.JwtKey)
}

func AuthenticateToken(reqToken string, claims *c.Claims) (int, error) {
	
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		return 1, nil
	}

	tknStr := strings.TrimSpace(splitToken[1])

	tkn, err := jwt.ParseWithClaims(tknStr, claims, JWTKey)
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return 2, err
		}
		return 3, err
	}

	if !tkn.Valid {
		return 2, err
	}

	return 0, nil
}

// take pass and return hashed pass
func HashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	c.Check(err)
	return string(hash)
}

func Verify(pass string, hashedPass string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPass), []byte(pass))
}

func JWTKey(token *jwt.Token) (interface{}, error) {
	return c.JwtKey, nil
}