package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *server) routes() {
	sd := "/static/"
	s.r.PathPrefix(sd).Handler(http.StripPrefix(sd, http.FileServer(http.Dir("."+sd))))

	// s.r.HandleFunc("/api/", s.handleAPI())

	s.r.HandleFunc("/", s.jsonDecor(s.handleIndex()))
	s.r.HandleFunc("/about", s.handleTemplate("About me", "navigation.html", "about.html", "footer.html", "base.html"))
	s.r.HandleFunc("/contact", s.handleTemplate("Constact me", "navigation.html", "contact.html", "footer.html", "base.html"))

	s.r.HandleFunc("/users", s.handleAdduser()).Methods("POST")
	s.r.HandleFunc("/users", s.jsonDecor(s.handleUsers)).Methods("GET")
	s.r.HandleFunc("/users/{id}", s.jsonDecor(s.handleUser())).Methods("GET")
	s.r.HandleFunc("/filterusers", s.jsonDecor(s.handleFilterUser)).Methods("GET")
	s.r.HandleFunc("/users/{id}", s.jsonDecor(s.handleUpdateUser)).Methods("PUT")
	s.r.HandleFunc("/users/{id}", s.jsonDecor(s.handleDeleteUser)).Methods("DELETE")
	s.r.HandleFunc("/password", s.handlePassword())

	s.r.HandleFunc("/status", s.handleStatus) // example how to use the simplest way
	s.r.HandleFunc("/admin", s.loginOnly(s.handleAdmin))
}

func (s *server) handleFilterUser(w http.ResponseWriter, r *http.Request) {

	log.Print("handleFilterUser start")
	defer log.Print("handleFilterUser end")
	// #### FIND OPTIONS #######
	options := options.FindOptions{}
	// Sort by `login` field descending
	// options.Sort = bson.D{{"login", -1}}//Composite literal uses unkeyed fields
	options.Sort = bson.D{primitive.E{Key: "login", Value: -1}}
	options.SetLimit(100)
	// options.SetSkip(1)//just for test

	// Request variables
	rvars := mux.Vars(r)
	for rvar := range rvars {
		log.Printf("Request variables %v", rvar)
	}

	// #### FILTER PARAMETERS #######
	params := r.URL.Query()

	// all params
	for param := range params {
		log.Printf("query params %+v ", param)
	}

	filter := bson.M{}
	var id primitive.ObjectID
	// ID requires object ID
	if _, ok := params["id"]; ok {
		id, _ = primitive.ObjectIDFromHex(params.Get("id"))
		filter["_id"] = id
	}
	if _, ok := params["login"]; ok {

		// filter = bson.M{"login": bson.M{"$regex": params.Get("login")}}//it works!!
		// filter["login"] = bson.M{"$regex": params.Get("login")}          //works but with case sensitivity

		// ##### SELECT * FROM USERS WHERE UPPER(LOGIN) LIKE %UPPER(SEARCH_TEXT)%

		filter["login"] = bson.M{"$regex": `(?i)` + params.Get("login")} //this magic works

		// it's a hack - just for test the range
		filter["age"] = bson.M{"$gte": 19, "$lte": 25}

	}

	// more sophisticated conditions - just for testing
	conditions := bson.M{"name": bson.M{"$regex": "me"},
		"$or": []bson.M{
			bson.M{"repair": bson.M{"$eq": "ac"}},
		},
		"$and": []bson.M{
			bson.M{"repair": bson.M{"$eq": "tv"}},
			bson.M{"phone": bson.M{"$gte": 1091, "$lte": 1100}},
		}}

	log.Printf("Conditions %+v", conditions)

	cur, err := s.db.Collection("users").Find(s.c, filter, &options)
	if err != nil {
		log.Printf("Problem ... %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cur.Close(s.c)

	var usrs []User
	// CURSOR HAVE NEXT METHOD
	for cur.Next(s.c) {
		var usr User
		err := cur.Decode(&usr)
		if err != nil {
			log.Printf("User problem err: %v\n", err)
			// continue // in case problematic are omitted
		}
		pass := usr.Password

		if err := bcrypt.CompareHashAndPassword([]byte(pass), []byte("testowe")); err == nil {
			usr.Password = "testowe" //just tu print unfriendly field
		}
		usrs = append(usrs, usr)
	}

	if err := cur.Err(); err != nil {
		log.Printf(" cur Problem ... %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	if err := json.NewEncoder(w).Encode(usrs); err != nil {
		log.Printf(" json Problem ... %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (s *server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := primitive.ObjectIDFromHex(params["id"])
	if err != nil {
		log.Printf("handleDeleteUser: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	var del *mongo.DeleteResult
	del, err = s.db.Collection("users").DeleteOne(s.c, filter)
	if err != nil {
		log.Printf("handleDeleteUser %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(del)

}

func (s *server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Problem ... %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// log.Printf("Buffer: %+v\n", user)

	if err := user.OK(); err != nil {
		log.Printf("Problem ... %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	params := mux.Vars(r)
	id, _ := primitive.ObjectIDFromHex(params["id"])

	// filter := bson.D{{"_id", id}} //this cause linnet problem
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	var result *mongo.UpdateResult
	// json.Unmarshal([]byte(`{ "$set": {"year": 1998}}`), &update)
	// update := bson.D{{"$set", &user}}
	update := bson.D{primitive.E{Key: "$set", Value: &user}}
	result, err := s.db.Collection("users").UpdateOne(s.c, filter, update)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	if err := json.NewEncoder(w).Encode(result); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		// http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (s *server) handleUser() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id, _ := primitive.ObjectIDFromHex(params["id"])
		// log.Printf("ID: %v\n", id)

		var usr User

		filter := bson.M{"_id": id}
		// it doest work for embedded struct
		// res := s.db.Collection("users").FindOne(s.c, User{ID: id})

		err := s.db.Collection("users").FindOne(s.c, filter).Decode(&usr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		if err := json.NewEncoder(w).Encode(usr); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (s *server) handleUsers(w http.ResponseWriter, r *http.Request) {

	cur, err := s.db.Collection("users").Find(s.c, bson.D{})
	if err != nil {
		log.Printf("Problem ... %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer cur.Close(s.c)

	var usrs []User

	for cur.Next(s.c) {
		var usr User
		err := cur.Decode(&usr)
		if err != nil {
			log.Printf("User problem err: %v\n", err)
			// continue // in case problematic are omitted
		}
		pass := usr.Password

		if err := bcrypt.CompareHashAndPassword([]byte(pass), []byte("testowe")); err == nil {
			usr.Password = "testowe"
		}
		usrs = append(usrs, usr)
	}
	if err := cur.Err(); err != nil {
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
	fmt.Fprintf(w, "<h1>Hi admin user, you are authorized to access this page!</h1>")

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
	return s.handleTemplate("Home page", "navigation.html", "home.html", "footer.html", "base.html")
}

// handleStatus - function presents OK answer if everything is ok
func (s *server) handleStatus(w http.ResponseWriter, r *http.Request) {
	// check if everyting is ok and set proper value of state
	// remember this is registering func and this before return is fired only once - like sync.Once
	state := "OK" //temporary everything is ok
	log.Print("request for status")
	fmt.Fprintf(w, state)

}

func (s *server) handleAdduser() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		last := TakeTime("handleAdduser")
		defer last()

		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Printf("Problem ... %v \n %+v\n", err, r.Body)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := user.OK(); err != nil {
			log.Printf("Problem ... %v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// log.Printf("Buffer: %+v\n", user)

		// rlt, err := ioutil.ReadAll(r.Body)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }
		// log.Printf("BODY: %+v\n", string(rlt))
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
		//passed func will be executed - it could be with conditions (here or decoration - just som action before/after)
		hf(w, r)

	}
}

func (s *server) jsonDecor(hf http.HandlerFunc) http.HandlerFunc {
	//here can be placed first time used code (like doOnce)

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("content-type", "application/json")
		hf(w, r)
	}

}

func (s *server) handleTemplate(title string, files ...string) http.HandlerFunc {

	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Print("handleTemplate start", files)
		defer log.Print("handleTemplate end")
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
