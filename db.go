package main

import (
	"database/sql"
	"fmt"
	"log"
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

// InsertPayment inserts a new payment record into the database
func InsertPayment(db *sql.DB, p *Payment) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var id int
	err = tx.QueryRow(`INSERT INTO payments (amount, currency, reference) VALUES ($1, $2, $3) RETURNING id`,
		p.Amount, p.Currency, p.Reference).Scan(&id)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

// GetPayment retrieves a payment record by ID from the database
func GetPayment(db *sql.DB, id int) (*Payment, error) {
	p := &Payment{}
	err := db.QueryRow(`SELECT id, amount, currency, reference, status, created_at FROM payments WHERE id = $1`, id).
		Scan(&p.ID, &p.Amount, &p.Currency, &p.Reference, &p.Status, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}
