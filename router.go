package main

import (
	"fmt"
	"log"
	"net/http"
)

func (s *server) routes() {

	// s.r.HandleFunc("/api/", s.handleAPI())
	// s.r.HandleFunc("/about", s.handleAbout())
	s.r.HandleFunc("/", s.handleIndex())

	s.r.HandleFunc("/status", s.handleStatus())
}

// handleStatus - function prensers OK answer if everything is ok
func (s *server) handleStatus() http.HandlerFunc {
	// chack if everyting is ok and set proper value of state
	// remember this is registring func and this before return is fired only once
	state := "OK" //temporary everything is ok
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("request for status")
		fmt.Fprintf(w, state)
	}
}

func (s *server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO return index page
		fmt.Fprintf(w, `<h1>Hi there - MONGODB playground</h1>
						<h2>Still under construction ...</h2>`)
	}
}

func (s *server) handleAdduser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
