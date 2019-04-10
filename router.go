package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

func (s *server) routes() {
	sd := "/static/"
	s.r.PathPrefix(sd).Handler(http.StripPrefix(sd, http.FileServer(http.Dir("."+sd))))

	// s.r.HandleFunc("/api/", s.handleAPI())

	s.r.HandleFunc("/", s.handleIndex())
	s.r.HandleFunc("/about", s.handleTemaplate("About me", "navigation.html", "about.html", "footer.html", "base.html"))
	s.r.HandleFunc("/contact", s.handleTemaplate("Constact me", "navigation.html", "contact.html", "footer.html", "base.html"))

	s.r.HandleFunc("/admin", s.loginOnly(s.handleAdmin()))

	s.r.HandleFunc("/users", s.handleAdduser()).Methods("POST")

	s.r.HandleFunc("/status", s.handleStatus())
}

func (s *server) handleAdmin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO add body
	}
}

func (s *server) handleIndex() http.HandlerFunc {
	//
	return s.handleTemaplate("Home page", "navigation.html", "home.html", "footer.html", "base.html")
}

// handleStatus - function prensers OK answer if everything is ok
func (s *server) handleStatus() http.HandlerFunc {
	// chack if everyting is ok and set proper value of state
	// remember this is registring func and this before return is fired only once - like sync.Once
	state := "OK" //temporary everything is ok
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("request for status")
		fmt.Fprintf(w, state)
	}
}

func (s *server) handleAdduser() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		log.Print("handleAdduser start")
		defer log.Print("handleAdduser end")

		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Printf("Problem ... %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Buffer: %+v\n", user)

		rlt, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("BODY: %+v\n", string(rlt))
		// take parameters
		params := mux.Vars(r)
		for param := range params {
			log.Print("Params: ", param)
		}

		w.WriteHeader(http.StatusOK)

		// if _, err := io.Copy(w, &buf); err != nil {
		// 	log.Print("respond:", err)
		// }

		usr, err := addUser(s.c, s.db, user, "users")
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
		log.Print("handlePassword start")
		defer log.Print("handlePassword end")

	}
}

func (s *server) loginOnly(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("Checking if logon")
		// TODO - implement condition
		condition := true
		if !condition {
			http.NotFound(w, r)
			return
		}
		hf(w, r)

	}
}

func (s *server) handleTemaplate(title string, files ...string) http.HandlerFunc {

	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("handleTemaplate start", files)
		defer log.Print("handleTemaplate end")
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
			"title": title,
		}
		// log.Printf("tpl.Tree.Root.String(): %s\n", tpl.Tree.Root.String())
		err := tpl.ExecuteTemplate(w, "base", data)
		if err != nil {
			log.Print("Problem with template", err)
		}

	}
}
