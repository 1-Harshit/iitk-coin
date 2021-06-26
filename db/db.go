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

	// Keep track of all transactions
	command = `CREATE TABLE IF NOT EXISTS "Transaction" (
		"tnxno"	INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE,
		"from"	INTEGER NOT NULL,
		"to"	INTEGER NOT NULL,
		"sent"	REAL NOT NULL DEFAULT 0,
		"tax"	REAL NOT NULL DEFAULT 0,
		"time"	DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		"remarks" TEXT,
		FOREIGN KEY("from") REFERENCES "User"("roll"),
		FOREIGN KEY("to") REFERENCES "User"("roll")
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

// Check if the user exists
func UserExists(roll int) bool {
	_, err := GetCoins(roll)
	return err != sql.ErrNoRows
}

// Get details present in user table
func GetUser(roll int) (c.User, error) {
	var usr c.User

	err := Mydb.QueryRow(`SELECT "roll", "name", "email", "password" FROM "main"."User" WHERE roll = $1`, roll).Scan(&usr.Roll, &usr.Name, &usr.Email, &usr.Password)
	if err != nil {
		return c.User{}, err
	} else {
		return usr, nil
	}
}

// Get details present in wallet table
func GetWallet(roll int) (c.Wallet, error) {
	var wal c.Wallet

	err := Mydb.QueryRow(`SELECT "roll", "coins", "usrtype", "batch" FROM "main"."Wallet" WHERE roll = $1`, roll).Scan(&wal.Roll, &wal.Coins, &wal.UsrType, &wal.Batch)
	if err != nil {
		return c.Wallet{}, err
	} else {
		return wal, nil
	}
}

func GetCoins(roll int) (float64, error) {

	var coins float64
	err := Mydb.QueryRow(`SELECT "coins" FROM "main"."Wallet" WHERE roll = $1`, roll).Scan(&coins)
	if err != nil {
		return -1, err
	} else {
		return coins, nil
	}
}

func RewardCoins(x c.Trnxn) error {

	tx, err := Mydb.Begin()
	if err != nil {
		return err
	}

	upd := `UPDATE "main"."Wallet" 
		SET coins= coins + $1 
		WHERE "roll"=$2 AND coins+$1<$3;`

	statement, err := tx.Prepare(upd)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer statement.Close()

	stmt, err := statement.Exec(x.Coins, x.To, c.MaxCoins)
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
		if UserExists(x.To) {
			tx.Rollback()
			return errors.New("max limit of coins for user exeeded")
		}
		tx.Rollback()
		return errors.New("no Such roll found")
	}
	txn_stm := `INSERT INTO "main"."Transaction"
		("from", "to", "sent", "remarks")
		VALUES (?, ?, ?, ?);`

	statement1, err := tx.Prepare(txn_stm)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer statement1.Close()

	_, err = statement1.Exec(x.From, x.To, x.Coins, x.Rem)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

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
	stmt, err = statement.Exec(t.Coins-tax, t.Roll, c.MaxCoins)
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

	_, err = statement1.Exec(f.Roll, t.Roll, t.Coins, tax, t.Rem)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

func EnoughTrans(roll int) bool {
	count := 0
	err := Mydb.QueryRow(`SELECT COUNT(*) FROM "main"."Transaction" WHERE to = $1`, roll).Scan(&count)
	if err != nil {
		return false
	}
	return count >= c.MinEvents
}
