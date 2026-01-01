package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

func setupAPI() *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.POST("/payments", createPayment)
	e.GET("/payments/:id", getPayment)
	return e
}

func createPayment(c echo.Context) error {
	// Parse request body
	p := new(Payment)
	if err := c.Bind(p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Validate input
	if err := validate.Struct(p); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Insert into database
	id, err := InsertPayment(db, p)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Reference must be unique or invalid data"})
	}

	err = PublishPaymentID(id)
	if err != nil {
		log.Printf("Publish error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Messaging error"})
	}

	// Return created payment
	p.ID = id
	p.Status = "PENDING"
	return c.JSON(http.StatusCreated, p)
}

func getPayment(c echo.Context) error {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
	}

	p, err := GetPayment(db, id)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Payment not found"})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database error"})
	}
	return c.JSON(http.StatusOK, p)
}
