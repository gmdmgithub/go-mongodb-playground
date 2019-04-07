package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// User - model for user
type User struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Login     string             `json:"login,omitempty" bson:"login,omitempty"`
	Password  string             `json:"password,omitempty" bson:"password,omitempty"`
	CreatedAt primitive.DateTime `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type server struct {
	db *mongo.Database
	r  *mux.Router
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Fatal problem during initialization: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("Welcome to the playground with mongoDB", os.Getenv("DB"))

	if err := run(); err != nil {
		log.Printf("Fatal problem during initialization: %v\n", err)
		os.Exit(1)
	}
}

func run() error {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	db, err := configDB(ctx)
	if err != nil {
		return err
	}

	s := &server{
		r:  mux.NewRouter(),
		db: db,
	}
	s.routes()

	p, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		log.Println("No http port in .env file, default 8000 taken")
		p = ":8000"
	}
	log.Printf("Server starts at port %s \n", p)
	return http.ListenAndServe(p, s.r)

}
