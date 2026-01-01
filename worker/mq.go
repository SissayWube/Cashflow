package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/streadway/amqp"
)

var paymentQueue string

func ConnectMQ() error {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found.")
		return err
	}

	paymentQueue = os.Getenv("MQ_QUEUE")

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

	q, err := mqChan.QueueDeclare(os.Getenv("MQ_QUEUE"), true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("Queue declare error: %v", err)
	}
	
	// Set QoS to prefetch 1 message for fair work distribution across workers
	err = mqChan.Qos(1, 0, false)
	if err != nil {
		log.Fatalf("QoS error: %v", err)
	}

	msgs, err = mqChan.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("Consume error: %v", err)
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
