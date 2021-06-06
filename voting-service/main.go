package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/streadway/amqp"
	"log"
	"net/http"
	"os"
)

var rabbit_host = os.Getenv("RABBIT_HOST")
var rabbit_port = os.Getenv("RABBIT_PORT")
var rabbit_user = os.Getenv("RABBIT_USERNAME")
var rabbit_password = os.Getenv("RABBIT_PASSWORD")

var l = log.New(os.Stdout, "voting-service", log.LstdFlags)


func main() {
	sm := mux.NewRouter()
	postR := sm.Methods(http.MethodPost).Subrouter()
	postR.HandleFunc("/{message}", submit)

	fmt.Println("Running...")
	l.Fatalln(http.ListenAndServe(":3000", sm))
}

func submit(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	msg := vars["message"]
	fmt.Println("Received message: " + msg)

	conn, err := amqp.Dial("amqp://" + rabbit_user + ":" + rabbit_password + "@" + rabbit_host + ":" + rabbit_port + "/" )

	if err != nil {
		l.Fatalf("%s: %s", "Failed to connect to RabbitMQ\n", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		l.Fatalf("%s: %s", "Failed to open a channel\n", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"publisher",
		false,
		false,
		false,
		false,
		nil,
		)

	if err != nil {
		l.Fatalf("%s:%s", "Failed to declear a queue", err)
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})


	if err != nil {
		l.Fatalf("%s:%s", "Failed to publish the message", err)
	}

	fmt.Println("published successfully")
}