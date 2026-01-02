package main

import (
	"fmt"
	"log"
	"os"

	"github.com/streadway/amqp"
)

var paymentQueue string

func ConnectMQ() error {
	paymentQueue = os.Getenv("MQ_QUEUE")
	if paymentQueue == "" {
		return fmt.Errorf("MQ_QUEUE environment variable not set")
	}

	var err error
	mqConn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:5672/",
		os.Getenv("MQ_USER"), os.Getenv("MQ_PASSWORD"), os.Getenv("MQ_HOST")))
	if err != nil {
		return fmt.Errorf("MQ connection error: %v", err)
	}

	mqChan, err = mqConn.Channel()
	if err != nil {
		return fmt.Errorf("MQ channel error: %v", err)
	}

	_, err = mqChan.QueueDeclare(paymentQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Queue declare error: %v", err)
	}

	log.Println("Connected to RabbitMQ successfully")
	return nil
}

func PublishPaymentID(id int) error {
	msg := amqp.Publishing{Body: []byte(fmt.Sprintf("%d", id))}
	err := mqChan.Publish("", paymentQueue, false, false, msg)
	if err != nil {
		return fmt.Errorf("publish error for payment %d: %v", id, err)
	}
	log.Printf("Published payment %d to queue", id)
	return nil
}
