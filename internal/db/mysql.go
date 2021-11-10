package database

import (
	// Built-ins
	"database/sql"
	"fmt"
	"log"
	"os"

	// External libs
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var Db *sql.DB

func InitDb() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	// Reading variables from .env
	var (
		dbUser    = os.Getenv("DB_USER") // e.g. 'my-db-user'
		dbPwd     = os.Getenv("DB_PASS") // e.g. 'my-db-password'
		dbTCPHost = os.Getenv("DB_HOST") // e.g. '127.0.0.1' ('172.17.0.1' if deployed to GAE Flex)
		dbPort    = os.Getenv("DB_PORT") // e.g. '3306'
		dbName    = os.Getenv("DB_NAME") // e.g. 'my-database'
	)
	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPwd, dbTCPHost, dbPort, dbName)

	// fmt.Println("DB LINK:", dbURI)

	db, err := sql.Open("mysql", dbURI)
	if err != nil {
		log.Panic(err)
	}

	// Test db connection
	if err := db.Ping(); err != nil {
		log.Panic(err)
	}

	// Successfully intialized connection to db
	Db = db
	fmt.Println("Connected to db")
}
