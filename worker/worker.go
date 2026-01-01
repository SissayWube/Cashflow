package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/streadway/amqp"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	// Connect to DB
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:5672/",
		os.Getenv("MQ_USER"), os.Getenv("MQ_PASSWORD"), os.Getenv("MQ_HOST")))
	if err != nil {
		log.Fatalf("MQ connection error: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("MQ channel error: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("payments", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Queue declare error: %v", err)
	}

	// Set QoS to prefetch 1 message for fair work distribution across workers
	err = ch.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("QoS error: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Consume error: %v", err)
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

	tx, err := db.Begin()
	if err != nil {
		log.Printf("DB begin error for payment %d: %v", id, err)
		return
	}
	defer tx.Rollback()

	// Lock row and check status
	var status string
	err = tx.QueryRow(`SELECT status FROM payments WHERE id = $1 FOR UPDATE`, id).Scan(&status)
	if err == sql.ErrNoRows {
		log.Printf("Payment %d not found", id)
		return
	} else if err != nil {
		log.Printf("DB query error for payment %d: %v", id, err)
		return
	}

	if status != "PENDING" {
		log.Printf("Skipping payment %d (already %s)", id, status)
		if err = tx.Commit(); err != nil {
			log.Printf("Commit error for payment %d: %v", id, err)
		}
		return
	}

	// Simulate processing
	time.Sleep(2 * time.Second)
	newStatus := "SUCCESS"
	if rand.Float32() < 0.3 {
		newStatus = "FAILED"
	}

	_, err = tx.Exec(`UPDATE payments SET status = $1 WHERE id = $2`, newStatus, id)
	if err != nil {
		log.Printf("Update error for payment %d: %v", id, err)
		return
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Commit error for payment %d: %v", id, err)
		return
	}

	log.Printf("Processed payment %d: %s", id, newStatus)
}