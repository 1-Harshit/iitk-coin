package db

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"

	c "github.com/1-Harshit/iitk-coin/config"
)

var dbpath string = c.DBpath
var Mydb *sql.DB

// initialize connection and tables
func init() {
	var err error
	Mydb, err = sql.Open("sqlite3", dbpath)
	c.Check(err)
	createTable()
}

// create table
func createTable() {

	// contains generic info about User
	command := `CREATE TABLE IF NOT EXISTS "User" (
		"roll"		INTEGER NOT NULL,
		"name"		TEXT NOT NULL,
		"email"		TEXT NOT NULL UNIQUE,
		"password"	TEXT NOT NULL,
		"createdat"	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY("roll")
	);`
	statement, err := Mydb.Prepare(command)
	c.Check(err)

	statement.Exec()

	// The Wallet information
	command = `CREATE TABLE IF NOT EXISTS "Wallet" (
		"sl"		INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"roll"		INTEGER NOT NULL UNIQUE,
		"coins"		REAL NOT NULL DEFAULT 0,
		"usrtype" 	INTEGER NOT NULL DEFAULT 0,
		"batch" 	INTEGER NOT NULL,
		FOREIGN KEY("roll") REFERENCES "User"("roll")
	);`

	statement, err = Mydb.Prepare(command)
	c.Check(err)

	statement.Exec()

	// The Reward information
	command = `CREATE TABLE IF NOT EXISTS "Reward" (
		"rewardID"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"time"		DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"roll"		INTEGER NOT NULL,
		"coins"		INTEGER NOT NULL,
		"status" 	INTEGER NOT NULL DEFAULT 1,
		"remarks" 	TEXT
	);`
	

	statement, err = Mydb.Prepare(command)
	c.Check(err)

	statement.Exec()

	// Keep track of all transactions
	command = `CREATE TABLE IF NOT EXISTS "Transactions" (
		"tnxNo"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"time"	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"from"	INTEGER NOT NULL,
		"to"	INTEGER NOT NULL,
		"sent"	REAL NOT NULL DEFAULT 0,
		"tax"	REAL NOT NULL DEFAULT 0,
		"remarks" TEXT,
		FOREIGN KEY("from") REFERENCES "User"("roll"),
		FOREIGN KEY("to") REFERENCES "User"("roll")
	);`


	statement, err = Mydb.Prepare(command)
	c.Check(err)

	statement.Exec()

	// The Item information
	command = `CREATE TABLE IF NOT EXISTS "Store" (
		"itemNo"		INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"name"			TEXT NOT NULL,
		"value"			INTEGER NOT NULL,
		"isavailable" 	INTEGER NOT NULL DEFAULT 1
	);`

	statement, err = Mydb.Prepare(command)
	c.Check(err)

	statement.Exec()
	
	// The Redeem information
	command = `CREATE TABLE IF NOT EXISTS "Redeem" (
		"redeemID"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"time"		DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"roll"		INTEGER NOT NULL,
		"itemNo"	INTEGER NOT NULL,
		"status" 	INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY("itemNo") REFERENCES "Store"("itemNo")
	);`

	statement, err = Mydb.Prepare(command)
	c.Check(err)

	statement.Exec()
	
	// OTP
	command = `CREATE TABLE IF NOT EXISTS "OTPInfo" (
		"sl"		INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"roll"		INTEGER NOT NULL,
		"otp"		TEXT NOT NULL,
		"time"		INTEGER NOT NULL,
		"isUsed"	INTEGER NOT NULL DEFAULT 0
	);`

	statement, err = Mydb.Prepare(command)
	c.Check(err)
	defer statement.Close()
	statement.Exec()
}

// Insert into db
func InsertIT(dt c.User) error {
	// begin transaction
	tx, err := Mydb.Begin()
	if err != nil {
		return err
	}

	// Insert in User
	insert_user := `INSERT INTO "main"."User"
		("roll", "name", "email", "password")
		VALUES (?, ?, ?, ?);`

	statement, err := tx.Prepare(insert_user)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = statement.Exec(dt.Roll, dt.Name, dt.Email, dt.Password)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert in Wallet
	insert_wal := `INSERT INTO "main"."Wallet"
		("roll", "batch")
		VALUES (?, ?);`

	statement, err = tx.Prepare(insert_wal)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer statement.Close()

	_, err = statement.Exec(dt.Roll, dt.Batch)
	if err != nil {
		tx.Rollback()
		return err
	}

	// commit changes
	err = tx.Commit()
	return err
}

// Reward a user
func RewardCoins(x c.Wallet) error {

	tx, err := Mydb.Begin()
	if err != nil {
		return err
	}

	upd := `UPDATE "main"."Wallet" 
		SET coins = coins + $1 
		WHERE "roll"=$2 AND coins+$1<$3;`

	statement, err := tx.Prepare(upd)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer statement.Close()

	stmt, err := statement.Exec(x.Coins, x.Roll, c.MaxCoins)
	if err != nil {
		tx.Rollback()
		return err
	}

	count, err2 := stmt.RowsAffected()
	if err2 != nil {
		tx.Rollback()
		return err2
	}
	if count == 0 {
		if UserExists(x.Roll) {
			tx.Rollback()
			return errors.New("max limit of coins for user exeeded")
		}
		tx.Rollback()
		return errors.New("no Such roll found")
	}

	// insert into reward
	redeem_stmt := `INSERT INTO "main"."Reward" 
		("roll", "coins", "remarks")
		VALUES (?, ?, ?)
	;`

	statement1, err := tx.Prepare(redeem_stmt)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer statement1.Close()

	_, err = statement1.Exec(x.Roll, x.Coins, x.Rem)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// Transfer coins to another user
func TransferCoins(t c.Wallet, f *c.Claims) error {
	if t.UsrType == 1 {
		return errors.New("frozen account")
	} else {
		if t.UsrType == 2 && f.UsrType != 1 {
			return errors.New("only Gensec and AH can transfer in this account")
		}
		if !(EnoughTrans(t.Roll) && EnoughTrans(f.Roll)) {
			return errors.New("not participated in enough events yet")
		}
	}

	tx, err := Mydb.Begin()
	if err != nil {
		return err
	}

	from_st := `UPDATE "main"."Wallet" 
		SET coins= coins - $1 
		WHERE "roll"=$2 AND coins>=$1;`

	statement, err := tx.Prepare(from_st)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer statement.Close()
	stmt, err := statement.Exec(t.Coins, f.Roll)
	if err != nil {
		tx.Rollback()
		return err
	}
	count, err2 := stmt.RowsAffected()
	if err2 != nil {
		tx.Rollback()
		return err2
	}
	if count == 0 {
		tx.Rollback()
		return errors.New("sender wallet doesn't enough capacity")
	}

	to_st := `UPDATE "main"."Wallet" 
		SET coins= coins + $1 
		WHERE "roll"=$2 AND coins+$1<$3;`

	statement1, err := tx.Prepare(to_st)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer statement1.Close()
	tax := c.CalculateTax(t, f)
	stmt, err = statement1.Exec(t.Coins-tax, t.Roll, c.MaxCoins)
	if err != nil {
		tx.Rollback()
		return err
	}

	count, err2 = stmt.RowsAffected()
	if err2 != nil {
		tx.Rollback()
		return err2
	}
	if count == 0 {
		tx.Rollback()
		return errors.New("max limit of coins for user exeeded")
	}

	txn_stm := `INSERT INTO "main"."Transaction"
		("from", "to", "sent", "tax", "remarks")
		VALUES (?, ?, ?, ?, ?);`

	statement2, err := tx.Prepare(txn_stm)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer statement2.Close()

	_, err = statement2.Exec(f.Roll, t.Roll, t.Coins, tax, t.Rem)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// if the user can transfer coins
func EnoughTrans(roll int) bool {
	count := 0
	err := Mydb.QueryRow(`SELECT COUNT(*) FROM "main"."Reward" WHERE roll = $1`, roll).Scan(&count)
	if err != nil {
		return false
	}
	return count >= c.MinEvents
}
