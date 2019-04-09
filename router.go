package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/gorilla/mux"
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
	return func() http.HandlerFunc {
		log.Println("handleIndex")
		return s.handleTemaplate("Home page", "navigation.html", "home.html", "footer.html", "base.html")
	}()
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

func (s *server) handleAdduser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Println("handleAdduser start")
		defer log.Println("handleAdduser end")

		rlt, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Panicln("BODY: ", rlt)
		// take parameters
		params := mux.Vars(r)
		for param := range params {
			log.Println("Params: ", param)
		}

		// var buf bytes.Buffer
		// if err := json.NewEncoder(&buf).Encode(data); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// w.WriteHeader(http.StatusOK)
		// if _, err := io.Copy(w, &buf); err != nil {
		// 	log.Println("respond:", err)
		// }

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

func (s *server) handleTemaplate(title string, files ...string) http.HandlerFunc {

	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("handleTemaplate start", files)
		defer log.Println("handleTemaplate end")
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
			log.Println("Problem with template", err)
		}

	}
}
