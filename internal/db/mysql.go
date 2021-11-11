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

func InitDb(testing ...bool) {
	env := []string{"DB_USER", "DB_PASS", "DB_HOST", "DB_PORT", "DB_NAME"}
	if len(testing) > 0 {
		env = []string{"DB_TEST_USER", "DB_TEST_PASS",
			"DB_TEST_HOST", "DB_TEST_PORT", "DB_TEST_NAME"}
	}

	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	// Reading variables from .env
	var (
		dbUser    = os.Getenv(env[0]) // e.g. 'my-db-user'
		dbPwd     = os.Getenv(env[1]) // e.g. 'my-db-password'
		dbTCPHost = os.Getenv(env[2]) // e.g. '127.0.0.1' ('172.17.0.1' if deployed to GAE Flex)
		dbPort    = os.Getenv(env[3]) // e.g. '3306'
		dbName    = os.Getenv(env[4]) // e.g. 'my-database'
	)
	dbURI := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPwd, dbTCPHost, dbPort, dbName)
	fmt.Println(dbURI)

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
	print := fmt.Sprintf("Connected to @%s:%s/%s", dbTCPHost, dbPort, dbName)
	fmt.Println(print)
}
