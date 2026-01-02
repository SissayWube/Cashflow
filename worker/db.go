package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// ConnectDB establishes a connection to the PostgreSQL database
func ConnectDB() (*sql.DB, error) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("DB open error: %v", err)
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		log.Printf("DB ping error: %v", err)
		return nil, err
	}

	log.Println("Connected to DB successfully")
	return db, nil
}

// UpdatePayment updates the payment status in the database
func UpdatePayment(db *sql.DB, paymentID int) error {

	tx, err := db.Begin()
	if err != nil {
		log.Printf("DB begin error for payment %d: %v", paymentID, err)
		return err
	}
	defer tx.Rollback()

	// Lock row and check status
	var status string
	err = tx.QueryRow(`SELECT status FROM payments WHERE id = $1 FOR UPDATE`, paymentID).Scan(&status)
	if err == sql.ErrNoRows {
		log.Printf("Payment %d not found", paymentID)
		return err
	} else if err != nil {
		log.Printf("DB query error for payment %d: %v", paymentID, err)
		return err
	}

	if status != "PENDING" {
		log.Printf("Skipping payment %d (already %s)", paymentID, status)
		return tx.Commit()
	}

	// Simulate processing (random success/fail)
	time.Sleep(2 * time.Second)
	newStatus := "SUCCESS"
	if rand.Float32() < 0.3 {
		newStatus = "FAILED"
	}

	_, err = tx.Exec(`UPDATE payments SET status = $1 WHERE id = $2`, newStatus, paymentID)
	if err != nil {
		log.Printf("Update error for payment %d: %v", paymentID, err)
		return err
	}

	if err = tx.Commit(); err != nil {
		log.Printf("Commit error for payment %d: %v", paymentID, err)
		return err
	}

	log.Printf("Processed payment %d: %s", paymentID, newStatus)
	return nil
}
