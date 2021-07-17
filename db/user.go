package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	c "github.com/1-Harshit/iitk-coin/config"
)

// Check if the user exists
func UserExists(roll int) bool {
	_, err := GetCoins(roll)
	return err != sql.ErrNoRows
}

// Get details present in user table
func GetUser(roll int) (c.User, error) {
	var usr c.User

	err := Mydb.QueryRow(`
		SELECT User.roll , User.name , User.email , User.password, Wallet.usrtype, Wallet.batch 
		FROM "main"."User" 
		INNER JOIN "main"."Wallet" ON User.roll = Wallet.roll
		WHERE User.roll = ?
	;`, roll).Scan(&usr.Roll, &usr.Name, &usr.Email, &usr.Password, &usr.UsrType, &usr.Batch)
		
	if err != nil {
		return c.User{}, err
	} 
	return usr, nil
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

// Get coins of a user
func GetCoins(roll int) (float64, error) {

	var coins float64
	err := Mydb.QueryRow(`SELECT "coins" FROM "main"."Wallet" WHERE roll = $1`, roll).Scan(&coins)
	if err != nil {
		return -1, err
	} else {
		return coins, nil
	}
}

// Get from Rewards
func GetUserRewards(roll int) ([]c.Reward, error) {
	command := `SELECT "time", "coins", "remarks" FROM "main"."Rewards" WHERE "roll" = $1 AND status = 1;`
	statement, err := Mydb.Prepare(command)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	var rewards []c.Reward
	rows, err := statement.Query(roll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var reward c.Reward
		err = rows.Scan(&reward.Time, &reward.Coins, &reward.Remarks)
		if err != nil {
			return nil, err
		}
		rewards = append(rewards, reward)
	}
	return rewards, nil
}

// Get from Transactions
func GetUserTransaction(roll int) ([]c.Tnxn, error) {
	command := `SELECT "time", "from", "to", "sent", "tax", "remarks" FROM "main"."Transactions" WHERE "to" = $1 OR from = $1;`
	statement, err := Mydb.Prepare(command)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	var transactions []c.Tnxn
	rows, err := statement.Query(roll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var transaction c.Tnxn
		err = rows.Scan(&transaction.Time, &transaction.From, &transaction.To, &transaction.Sent, &transaction.Tax, &transaction.Remarks)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func GetUserRedeem(roll int) ([]c.Redeem, error) {
	command := `
		SELECT Redeem.sl, Redeem.time, Redeem.roll, Redeem.status, Redeem.itemNo, Store.name, Store.value
		FROM "main"."Redeem" 
		INNER JOIN "main"."Store" ON Redeem.itemNo=Store.itemNo
		WHERE Redeem.roll = $1 
		ORDER BY Redeem.time
	;`
	statement, err := Mydb.Prepare(command)
	if err != nil {
		return nil, err
	}
	defer statement.Close()
	var redeem []c.Redeem
	rows, err := statement.Query(roll)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var redeemItem c.Redeem
		err = rows.Scan(&redeemItem.Id, &redeemItem.Time, &redeemItem.Roll, &redeemItem.Status, &redeemItem.ItemNo, &redeemItem.Name , &redeemItem.Value)
		if err != nil {
			return nil, err
		}
		redeem = append(redeem, redeemItem)
	}
	return redeem, nil

}
