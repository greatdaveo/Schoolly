package sqlconnect

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() (*sql.DB, error) {
	fmt.Println("📍-------- Connecting to Database... ⏳")

	// err := godotenv.Load()
	// if err != nil {
	// 	return nil, err
	// }

	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	DB_PORT := os.Getenv("DB_PORT")
	host := os.Getenv("HOST")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, DB_PORT, dbname)
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		// panic(err)
		return nil, err
	}

	fmt.Println("✅ Connected to DATABASE!!!")

	return db, nil
}
