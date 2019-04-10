package main

import (
	"context"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type server struct {
	db *mongo.Database
	r  *mux.Router
	c  context.Context
}
