package main

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"os"
	"time"
)

var l = log.New(os.Stdout, "worker", log.LstdFlags)

var rabbit_host = os.Getenv("RABBIT_HOST")
var rabbit_port = os.Getenv("RABBIT_PORT")
var rabbit_user = os.Getenv("RABBIT_USERNAME")
var rabbit_password = os.Getenv("RABBIT_PASSWORD")

type Vote struct {
	ID primitive.ObjectID   `json:"_id,omitempty" bson:"_id, omitempty"`
	IDD int					`json:"idd,omitempty" bson:"idd, omitempty"`
	Cat int  			`json:"cat,omitempty" bson:"cat, omitempty"`
	Dog int				`json:"dog,omitempty" bson:"dog, omitempty"`

}

func main() {
	consume()
}

func consume() {

	conn, err := amqp.Dial("amqp://" + rabbit_user + ":" + rabbit_password + "@" + rabbit_host + ":" + rabbit_port + "/" )

	if err != nil {
		log.Fatalf("%s: %s\n", "Failed to connect to RabbitMQ\n", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("%s: %s\n", "Failed to open a channel\n", err)
	}

	q, err := ch.QueueDeclare(
		"publisher",
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("%s:%s\n", "Failed to declear a queue", err)
	}

	fmt.Println("Channel and the queue has been established")

	defer conn.Close()
	defer ch.Close()

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
		)



	if err != nil {
		log.Fatalf("%s:%s\n", "Failed to register consumer", err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received %s\n", d.Body)
			d.Ack(false)
			go updateDb(d.Body)
		}
	}()

	fmt.Println("Running...")

	<-forever
}

func updateDb(msg []byte) {

	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, _ := mongo.Connect(ctx, clientOptions)


	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		l.Println(err)
		return
	}

	l.Println("Successfully connected and pinged.")

	db := client.Database("votes")
	collection := db.Collection("vote")

	vote, err := findDb(collection)
	if err !=  nil {
		l.Println("couldn't find vote to update ", err)
		return
	}

	if string(msg) == "cat" {
		vote.Cat = vote.Cat + 1
		updateCat(collection, vote)
	}
	if string(msg) == "dog" {
		vote.Dog = vote.Dog + 1
		updateDog(collection, vote)
	}
}



func findDb(collection *mongo.Collection) (*Vote, error){

	var vote Vote

	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	err := collection.FindOne(ctx, bson.M{"idd": 1}).Decode(&vote)
	if err != nil {
		l.Println(err)
		return nil, err
	}
	return &vote, nil
}

func updateCat(collection *mongo.Collection, vote *Vote) {


		ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancle()
		_, err := collection.UpdateOne(ctx,
			bson.M{"idd": 1},
			bson.D{
				{"$set", bson.D{{"cat", vote.Cat}}},
			})
		if err != nil {
			l.Println(err)
		}

}

func updateDog(collection *mongo.Collection, vote *Vote) {

		ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancle()
		_, err := collection.UpdateOne(ctx,
			bson.M{"idd": 1},
			bson.D{
				{"$set", bson.D{{"dog", vote.Dog}}},
			})
		if err != nil {
			l.Println(err)
		}
}