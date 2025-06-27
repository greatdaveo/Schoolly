package sqlconnect

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB(dbname string) (*sql.DB, error) {
	fmt.Println("📍-------- Connecting to Database... ⏳")

	connectionString := "root:root@tcp(127.0.0.1:3306)/" + dbname
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		// panic(err)
		return nil, err
	}

	fmt.Println("✅ Connected to DATABASE!!!")

	return db, nil
}
