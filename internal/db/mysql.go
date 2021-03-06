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

// The first index specifies the env to read from
// including "development" and "local".
// If second string exists and equal to "testing",
// the function will never connect to the db.
//
// Default to production env with no/invalid inputs
func InitDb(stage ...string) {
	env := []string{"DB_USER", "DB_PASS", "DB_HOST", "DB_PORT", "DB_NAME"}
	if len(stage) > 0 {
		if stage[0] == "development" {
			env = []string{"DB_TEST_USER", "DB_TEST_PASS",
				"DB_TEST_HOST", "DB_TEST_PORT", "DB_TEST_NAME"}
		} else if stage[0] == "local" {
			env = []string{"DB_LOCAL_USER", "DB_LOCAL_PASS",
				"DB_LOCAL_HOST", "DB_LOCAL_PORT", "DB_LOCAL_NAME"}
		}
	}

	// Testing connection; never actually connect to db
	testing := false
	if len(stage) > 1 {
		if stage[1] == "testing" {
			testing = true
		}
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

	db, err := sql.Open("mysql", dbURI)
	if err != nil {
		log.Panic(err)
	}

	// Test db connection
	if err := db.Ping(); err != nil {
		log.Panic(err)
	}

	// Successfully intialized connection to db
	if !testing {
		Db = db
	}

	print := fmt.Sprintf("Connected to @%s:%s/%s", dbTCPHost, dbPort, dbName)
	fmt.Println(print)
}
