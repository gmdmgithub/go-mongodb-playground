package main

import (
	"context"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type server struct {
	db *mongo.Database
	r  *mux.Router
	c  context.Context
}

// TakeTime - util func printing duration of the func
func TakeTime(funcName string) func() {
	t := time.Now()
	log.Printf("Starts func %s", funcName)
	return func() {
		log.Printf("Func %s took time: %v", funcName, time.Since(t))
	}

}
