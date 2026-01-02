package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

var db *sql.DB
var mqConn *amqp.Connection
var mqChan *amqp.Channel
var msgs <-chan amqp.Delivery

func main() {
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

	// Connect to DB
	var err error
	db, err = ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Connect to MQ
	if err = ConnectMQ(); err != nil {
		log.Fatalf("Failed to connect to MQ: %v", err)
	}
	defer mqConn.Close()
	defer mqChan.Close()

	log.Println("Worker started, waiting for messages...")
	for d := range msgs {
		processPayment(d.Body)
		d.Ack(false) // Ack only after processing
	}
}

func processPayment(body []byte) {
	idStr := string(body)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid payment ID: %s", idStr)
		return
	}

	err = UpdatePayment(db, id)
	if err != nil {
		log.Printf("Error updating payment %d: %v", id, err)
		return
	}

}
