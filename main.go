package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/gorilla/mux"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Fatal problem duting initialization: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("Welcome to the playground with mongoDB", os.Getenv("DB"))

	if err := run(); err != nil {
		log.Printf("Fatal problem duting initialization: %v\n", err)
		os.Exit(1)
	}
}

func run() error {

	r := mux.NewRouter()

	r.Handle("/status", http.HandlerFunc(handleStatus))

	return http.ListenAndServe(":8080", r)

}

// handleStatus - function prensers OK answer if everything is ok
func handleStatus(w http.ResponseWriter, r *http.Request) {
	log.Println("request for status")
	fmt.Fprintf(w, "OK")
}
