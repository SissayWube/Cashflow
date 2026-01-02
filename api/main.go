package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var db *sql.DB
var mqConn *amqp.Connection
var mqChan *amqp.Channel

func main() {
	// Load .env only in non-production environments
	// Only load .env in local development (default to dev if APP_ENV not set)
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" || appEnv == "development" {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Could not load .env file: %v (using system env vars)", err)
		} else {
			log.Println(".env file loaded for local development")
		}
	} else {
		log.Println("Production mode: using container-injected environment variables")
	}

	var err error

	// connect
	db, err = ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// run migrations
	if err := MigrateDB(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	if err = ConnectMQ(); err != nil {
		log.Fatalf("Failed to connect to MQ: %v", err)
	}

	defer mqConn.Close()
	defer mqChan.Close()

	// setup the Echo server
	e := setupAPI()
	log.Println("API started on :8080")
	// start the server
	e.Logger.Fatal(e.Start(":8080"))
}
