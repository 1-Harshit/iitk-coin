package db

import (
	// "database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"

	c "github.com/1-Harshit/iitk-coin/config"
)

// Store otp in database
func StoreOTP(t c.User) string {
	expired := time.Now().Add(6 * time.Minute).Unix()

	command := `INSERT INTO "main"."OTPInfo" ("roll", "otp", "time")
		VALUES (?, ?, ?)
	;`
	statement, err := Mydb.Prepare(command)
	if err != nil {
		return err.Error()
	}
	defer statement.Close()
	_, err = statement.Exec(t.Roll, t.OTP, expired)
	if err != nil {
		return err.Error()
	}
	return ""
}

// get otp from database
func GetOTP(roll int) (string, int, error) {
	command := `SELECT "sl", "otp", "time", "isUsed" FROM "main"."OTPInfo" WHERE "roll" = ? ORDER BY "time" DESC;`
	statement, err := Mydb.Prepare(command)
	if err != nil {
		return "", -1, err
	}
	defer statement.Close()
	var otp string
	var timeout int64
	var isUsed int
	var sl int
	err = statement.QueryRow(roll).Scan(&sl, &otp, &timeout, &isUsed)
	if err != nil {
		return "", -1, err
	}
	if isUsed != 0 {
		return "", -1, errors.New("please generate a new otp")
	}
	if time.Now().Unix() > timeout {
		return "", -1, errors.New("no OTP found")
	}
	return otp, sl, nil
}

// count no of otp
func ExceedMaxOTP(roll int) bool {
	timestr := time.Now().Add(-5 * time.Minute).Unix()
	command := `SELECT COUNT(*) FROM "main"."OTPInfo" WHERE "roll" = ? AND isUsed=0 AND time > ?;`

	statement, err := Mydb.Prepare(command)
	if err != nil {
		return false
	}
	defer statement.Close()
	var count int
	err = statement.QueryRow(roll, timestr).Scan(&count)
	if err != nil {
		return false
	}
	return count > 2
}

// Mark otp as used
func MarkOTP(sl int) error {
	// Update OTP ID in database
	command := `UPDATE "main"."OTPInfo" SET "isUsed" = 1 WHERE "sl" = ?;`
	statement, err := Mydb.Prepare(command)
	if err != nil {
		return err
	}
	_, err = statement.Exec(sl)
	return err
}
