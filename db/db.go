package db

import (
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	
	c "github.com/1-Harshit/iitk-coin/config"
)

var dbpath = "./data.db"
var Mydb *sql.DB

func init() {
	c.Out("let's Go!")
	var err error
	Mydb, err = sql.Open("sqlite3", dbpath)
	check(err)
	createTable()
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
	c.Out("Tables created")
}

func InsertIT(dt c.User) error {

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

func GetUser(roll int) (c.User, error) {
	var usr c.User

	err := Mydb.QueryRow(`SELECT "roll", "name", "email", "password" FROM "main"."User" WHERE roll = $1`, roll).Scan(&usr.Roll, &usr.Name, &usr.Email, &usr.Password)
	if err != nil {
		return c.User{}, err
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

func RewardCoins(x c.Wallet) error {

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

func TransferCoins(t c.Trnxn) error {

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

func check(e error) {
	if e != nil {
		panic(e)
	}
}