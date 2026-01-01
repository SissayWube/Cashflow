package main

import (
	"database/sql"
	"log"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
)

var db *sql.DB
var mqConn *amqp.Connection
var mqChan *amqp.Channel
var msgs <-chan amqp.Delivery

func main() {
	// Connect to DB
	var err error
	db, err = ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Connect to MQ
	if err = ConnectMQ(); err != nil {
		log.Fatalf("Failed to connect to MQ: %v", err)
	}

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
