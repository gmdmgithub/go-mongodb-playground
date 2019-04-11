package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *server) routes() {
	sd := "/static/"
	s.r.PathPrefix(sd).Handler(http.StripPrefix(sd, http.FileServer(http.Dir("."+sd))))

	// s.r.HandleFunc("/api/", s.handleAPI())

	s.r.HandleFunc("/", s.decor(s.handleIndex()))
	s.r.HandleFunc("/about", s.handleTemaplate("About me", "navigation.html", "about.html", "footer.html", "base.html"))
	s.r.HandleFunc("/contact", s.handleTemaplate("Constact me", "navigation.html", "contact.html", "footer.html", "base.html"))

	s.r.HandleFunc("/users", s.handleAdduser()).Methods("POST")
	s.r.HandleFunc("/users", s.decor(s.handleUsers)).Methods("GET")
	s.r.HandleFunc("/passwd", s.handlePassword())

	s.r.HandleFunc("/status", s.handleStatus) // example how to use the simplest way
	s.r.HandleFunc("/admin", s.loginOnly(s.handleAdmin))
}

func (s *server) handleUsers(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("content-type", "application/json")

	coll, err := s.db.Collection("users").Find(s.c, bson.M{})
	if err != nil {
		log.Printf("Problem ... %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer coll.Close(s.c)

	var usrs []User

	for coll.Next(s.c) {
		var usr User
		err := coll.Decode(&usr)
		if err != nil {
			log.Printf("User problem :%T %+v err: %v\n", usr, usr, err)
			// continue
		}
		pass := usr.Password

		if err := bcrypt.CompareHashAndPassword([]byte(pass), []byte("testowe")); err == nil {
			usr.Password = "testowe"
		}
		_, err = json.Marshal(usr.CreatedAt)
		if err != nil {
			usr.CreatedAt = primitive.DateTime(time.Now().UnixNano() / int64(time.Millisecond))
			log.Printf("Date problem :%T %v\n", usr.CreatedAt, usr.CreatedAt)
		}
		usrs = append(usrs, usr)
	}
	if err := coll.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	// log.Printf("Users: %+v", usrs)
	if err := json.NewEncoder(w).Encode(usrs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (s *server) handleAdmin(w http.ResponseWriter, r *http.Request) {

	// TODO add body
	fmt.Fprintf(w, "<h1>Hi admin user, you are authorised to access this page!</h1>")

}

func (s *server) handlePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("handlePassword start")
		defer log.Print("handlePassword end")
		fmt.Fprintf(w, "Password set")
	}
}

func (s *server) handleIndex() http.HandlerFunc {
	//cover path
	return s.handleTemaplate("Home page", "navigation.html", "home.html", "footer.html", "base.html")
}

// handleStatus - function prensers OK answer if everything is ok
func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	// chack if everyting is ok and set proper value of state
	// remember this is registring func and this before return is fired only once - like sync.Once
	state := "OK" //temporary everything is ok
	log.Print("request for status")
	fmt.Fprintf(w, state)

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

func (s *server) loginOnly(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("Checking if logon")
		// TODO - implement condition
		condition := true
		if !condition {
			http.NotFound(w, r)
			return
		}
		//passed func will be executed - it could be with conditions (here or decoration - just som akction before/after)
		hf(w, r)

	}
}

func (s *server) decor(hf http.HandlerFunc) http.HandlerFunc {
	//here can be placed first time used code (like doOnce)

	return func(w http.ResponseWriter, r *http.Request) {

		log.Print("Here you can do some code before ......")
		hf(w, r)
		log.Print("Then you can start after part")
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
