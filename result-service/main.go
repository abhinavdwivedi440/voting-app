package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
	"os"
	"time"
)

type Vote struct {
	ID primitive.ObjectID   `json:"_id,omitempty" bson:"_id, omitempty"`
	IDD int					`json:"idd,omitempty" bson:"idd, omitempty"`
	Cat int  			`json:"cat,omitempty" bson:"cat, omitempty"`
	Dog int				`json:"dog,omitempty" bson:"dog, omitempty"`

}

var client *mongo.Client

var l = log.New(os.Stdout, "result-service", log.LstdFlags)

func getVotes(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "text/plain")
	var vote Vote


	db := client.Database("votes")


	collection := db.Collection("vote")

	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	err := collection.FindOne(ctx, bson.M{"idd": 1}).Decode(&vote)
	if err != nil {
		l.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = fmt.Fprintf(w, "Cats: %d\t\tDogs: %d\n", vote.Cat, vote.Dog)
	if err != nil {
		l.Fatalln(err)
	}
	return
}
func main() {


	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	clientOptions := options.Client().ApplyURI("mongodb://mongodb:27017")
	client, _ = mongo.Connect(ctx, clientOptions)


	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		l.Fatalln(err)
	}

	l.Println("Successfully connected and pinged.")

	sm := mux.NewRouter()
	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/", getVotes)

	l.Fatalln(http.ListenAndServe(":4000", sm))
}