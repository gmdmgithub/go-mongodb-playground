package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gmdmgithub/mongodb-first/config"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Version -- general struct to kep data about version and audit data
type Version struct {
	VerTag     string             `json:"ver_tag,omitempty" bson:"ver_tag,omitempty"`
	Created    primitive.DateTime `json:"created,omitempty" bson:"created,omitempty"`
	Updated    time.Time          `json:"updated,omitempty" bson:"updated,omitempty"`
	UsrCreated primitive.ObjectID `json:"usr_created" bson:"usr_created,omitempty"`
	UsrUpdated primitive.ObjectID `json:"usr_updated,omitempty" bson:"usr_updated,omitempty"`
}

// User - model for user
type User struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Login    string             `json:"login,omitempty" bson:"login,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
	Age      int                `json:"age,omitempty" bson:"age,omitempty"`
	Version  Version            `json:"version,omitempty" bson:"version,omitempty"`
	// CreatedAt primitive.DateTime `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

// OK - check if user data are ok
func (u *User) OK() error {

	if u.Login == "" || u.Password == "" {
		return errors.New("login or passward are required")
	}
	return nil
}

// Valid - general interface to check if struct object are ok to pass them to action
type Valid interface {
	OK() error
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Fatal problem during initialization: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("Welcome to the playground with mongoDB \u2318")

	if err := run(); err != nil {
		log.Printf("Fatal problem during initialization: %v\n", err)
		os.Exit(1)
	}
}

func run() error {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	config.LoadLog()

	db, err := configDB(ctx)
	if err != nil {
		return err
	}

	s := &server{
		r:  mux.NewRouter().StrictSlash(true),
		db: db,
		c:  ctx,
	}
	s.routes()

	p, ok := os.LookupEnv("HTTP_PORT")
	if !ok {
		log.Print("No http port in .env file, default 8000 taken")
		p = ":8000"
	}
	log.Printf("Server starts at port %s \n", p)
	return http.ListenAndServe(p, s.r)

}
