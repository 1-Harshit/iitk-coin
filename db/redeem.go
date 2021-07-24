package db

import (
	"errors"

	_ "github.com/mattn/go-sqlite3"

	c "github.com/1-Harshit/iitk-coin/config"
)

func GetItem(t int) (c.Item, error) {
	var res c.Item
	err := Mydb.QueryRow(`SELECT "name", "value" FROM "main"."Store" WHERE itemNo = $1`, t).Scan(&res.Name, &res.Value)
	if err != nil {
		return res, err
	}
	res.ItemNo = t
	return res, nil
}

func GetItems() ([]c.Item, error) {
	// query
	row, err := Mydb.Query(`SELECT "itemNo", "name", "value" FROM "main"."Store" WHERE isavailable=1 ORDER BY value`)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var AllItems []c.Item
	// loop thu the records
	for row.Next() {
		entry := c.Item{}
		err = row.Scan(&entry.ItemNo, &entry.Name, &entry.Value)
		if err != nil {
			return nil, err
		}
		AllItems = append(AllItems, entry)
	}
	if err := row.Err(); err != nil {
		return nil, err
	}
	return AllItems, nil
}

// Insert into db
func DeleteItems(t c.Item) error {

	// Insert in Store
	insert_item := `UPDATE "main"."Store" SET "isavailable"=0 WHERE "itemNo"=?;`

	statement, err := Mydb.Prepare(insert_item)
	if err != nil {
		return err
	}

	_, err = statement.Exec(t.ItemNo)

	return err
}

func InsertItems(t c.Item) error {

	// Insert in Store
	insert_item := `INSERT INTO "main"."Store" ("name", "value")
		VALUES (?, ?)
	;`

	statement, err := Mydb.Prepare(insert_item)
	if err != nil {
		return err
	}

	_, err = statement.Exec(t.Name, t.Value)

	return err
}

func ReqRedeem(t c.Redeem) error {
	row, err := Mydb.Query(`
		SELECT Redeem.redeemID, Redeem.itemNo, Store.value
		FROM "main"."Redeem" 
		INNER JOIN "main"."Store" ON Redeem.itemNo=Store.itemNo
		WHERE Redeem.status=0 AND Redeem.roll = ?
		ORDER BY Redeem.time
	;`, t.Roll)

	if err != nil {
		return err
	}
	defer row.Close()
	total := 0
	// loop thu the records
	for row.Next() {
		entry := c.Redeem{}
		err = row.Scan(&entry.Id, &entry.ItemNo, &entry.Value)
		if err != nil {
			return err
		}
		total += entry.Value
	}
	if err := row.Err(); err != nil {
		return err
	}
	itm, _ := GetItem(t.ItemNo)
	coins, _ := GetCoins(t.Roll)
	if int(coins) < total+itm.Value {
		return errors.New("not enough coins")
	}

	// Insert in Store
	stm := `INSERT INTO "main"."Redeem" ("roll", "itemNo")
		VALUES (?, ?)
	;`

	statement, err := Mydb.Prepare(stm)
	if err != nil {
		return err
	}

	_, err = statement.Exec(t.Roll, t.ItemNo)

	return err
}

func GetReedem() ([]c.Redeem, error) {
	// query
	row, err := Mydb.Query(`
		SELECT Redeem.redeemID, Redeem.time, Redeem.roll, Redeem.itemNo, Store.name, Store.value
		FROM "main"."Redeem" 
		INNER JOIN "main"."Store" ON Redeem.itemNo=Store.itemNo
		WHERE Redeem.status=0 
		ORDER BY Redeem.time
	;`)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	var AllItems []c.Redeem
	// loop thu the records
	for row.Next() {
		entry := c.Redeem{}
		err = row.Scan(&entry.Id, &entry.Time, &entry.Roll, &entry.ItemNo, &entry.Name, &entry.Value)
		if err != nil {
			return nil, err
		}
		AllItems = append(AllItems, entry)
	}
	if err := row.Err(); err != nil {
		return nil, err
	}
	return AllItems, nil
}

func RejectRedeem(t int) error {

	statement, err := Mydb.Prepare(`UPDATE "main"."Redeem" SET "status"=-1 WHERE "redeemID"=?;`)
	if err != nil {
		return err
	}

	_, err = statement.Exec(t)

	return err
}

func ApproveRedeem(t int) error {

	row := Mydb.QueryRow(`
		SELECT Redeem.redeemID, Redeem.time, Redeem.roll, Redeem.itemNo, Store.name, Store.value
		FROM "main"."Redeem" 
		INNER JOIN "main"."Store" ON Redeem.itemNo=Store.itemNo
		WHERE Redeem.status=0 AND Redeem.redeemID = ?
	;`, t)
	redm := c.Redeem{}
	err := row.Scan(&redm.Id, &redm.Time, &redm.Roll, &redm.ItemNo, &redm.Name, &redm.Value)
	if err != nil {
		return err
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

	stmt, err := statement.Exec(redm.Value, redm.Roll)
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

	// update reedeem status
	stm := `UPDATE "main"."Redeem" SET "status"=1 WHERE "redeemID"=?;`
	statement, err = tx.Prepare(stm)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = statement.Exec(redm.Id)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}
