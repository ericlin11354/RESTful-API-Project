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

func InitDb(envNames ...string) {
	// Establish connection with db
	if len(envNames) == 0 { // use default path (.env)
		if err := godotenv.Load(); err != nil {
			log.Fatal(err)
		}
	} else {
		for _, path := range envNames { // looping until
			fmt.Println("Path: ", path)
			if err := godotenv.Load(path); err != nil {
				continue
			} else {
				break
			}
		}
		log.Fatal("No valid .env found")
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
}
