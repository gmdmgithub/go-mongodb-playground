package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

func (s *server) routes() {
	sd := "/static/"
	s.r.PathPrefix(sd).Handler(http.StripPrefix(sd, http.FileServer(http.Dir("."+sd))))
	// set static files to be serve from static dir
	// s.r.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./"))))

	// s.r.HandleFunc("/api/", s.handleAPI())
	// s.r.HandleFunc("/about", s.handleAbout())
	s.r.HandleFunc("/", s.handleIndex())

	s.r.HandleFunc("/about", s.handleTemaplate("about.html", "base.html"))

	s.r.HandleFunc("/status", s.handleStatus())
	s.r.HandleFunc("/admin", s.loginOnly(s.handleAdmin()))
	s.r.HandleFunc("/users", s.handleAdduser()).Methods("POST")
}

func (s *server) handleAdmin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO add body
	}
}

// handleStatus - function prensers OK answer if everything is ok
func (s *server) handleStatus() http.HandlerFunc {
	// chack if everyting is ok and set proper value of state
	// remember this is registring func and this before return is fired only once - like sync.Once
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
		log.Println("handleAdduser start")
		defer log.Println("handleAdduser end")
		// first add as object
		user := User{
			Login:    "test 3",
			Password: "best",
		}

		usr, err := addUser(s.db, user, "users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if err := json.NewEncoder(w).Encode(&usr); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (s *server) handlePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("handlePassword start")
		defer log.Println("handlePassword end")

	}
}

func (s *server) loginOnly(loged http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Checking if logon")
		// TODO - implement condition
		condition := true
		if !condition {
			http.NotFound(w, r)
			return
		}
		loged(w, r)

	}
}

func (s *server) handleTemaplate(files ...string) http.HandlerFunc {

	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("handleTemaplate start")
		init.Do(func() {
			for i, file := range files {
				files[i] = filepath.Join("templates", file)
			}
			tpl, err = template.ParseFiles(files...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		})
		// there was a problem with template file (init func)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		// TODO execute the template

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		//get some data - temporary some static
		data := map[string]string{
			"name":  "Alex",
			"age":   "34",
			"title": "About me",
		}
		// log.Printf("tpl.Tree.Root.String(): %s\n", tpl.Tree.Root.String())
		err := tpl.ExecuteTemplate(w, "base", data)
		if err != nil {
			log.Println("Problem with template", err)
		}

	}
}
