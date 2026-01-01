package main

import "time"

type Payment struct {
	ID        int       `json:"id"`
	Amount    float64   `json:"amount" validate:"required,gt=0"`
	Currency  string    `json:"currency" validate:"required,oneof=ETB USD"`
	Reference string    `json:"reference" validate:"required"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
