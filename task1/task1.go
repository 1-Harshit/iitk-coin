package main

import (
	"database/sql"
	// "log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var dbpath = "./task1/data.db"
// var out = log.Println // alias cuz why not :-/

var out = nothing
func nothing(v ...interface{}) {
	return
}

type user struct {
	roll int
	name string
}

func main() {

	out("let's go!")

	// do else: panic: table User already exists
	os.Remove(dbpath)
	file, err := os.Create(dbpath)
	check(err)
	file.Close()
	out("data.db created")

	// initiate my db
	mydb, err := sql.Open("sqlite3", dbpath)
	check(err)
	defer mydb.Close() // procrastination on purpose aka defer

	// make the table 'User'
	createTable(mydb)

	// insert name of all idiots
	feedData(mydb, 200433, "Gawd of typos")
	feedData(mydb, 200070, "debate karne do")
	feedData(mydb, 200076, "excuse me brother")
	// will not be inserted
	feedData(mydb, 200000, "lemme sleep")
	feedData(mydb, 170000, "Random person")

	// confirm their existence
	display(mydb)
	out("Done! :)")
}

func createTable(db *sql.DB) {

	// meko sql aata hai :)
	command := `CREATE TABLE User (
		"rollno" integer NOT NULL PRIMARY KEY,		
		"name" varchar	
	  );`
	// check the command
	statement, err := db.Prepare(command)
	check(err)

	// do it already!
	statement.Exec()
	out("User table created")
}

func feedData(db *sql.DB, roll int, name string) {
	// // validate data if u wanna, say only ug-y20 roll
	// if roll > 200000 && roll < 201500 {
		dt := user{}
		dt.name = name
		dt.roll = roll
		insertIT(db, dt)
	// } else {
	// 	out("Only y20 plij")
	// }
}

func insertIT(db *sql.DB, dt user) {

	// why comment a simple sql :p
	insert_User := `INSERT INTO User (rollno, name) VALUES (?, ?)`

	statement, err := db.Prepare(insert_User)
	check(err)

	_, err = statement.Exec(dt.roll, dt.name)
	check(err)
}

func display(db *sql.DB) {
	// query
	row, err := db.Query("SELECT * FROM User ORDER BY rollno")
	check(err)
	defer row.Close()

	// loop thu the records
	for row.Next() {
		var roll int
		var name string
		row.Scan(&roll, &name)
		out("User: ", roll, " ", name)
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}