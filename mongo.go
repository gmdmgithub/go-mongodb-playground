package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/rs/zerolog/log"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type key string

const (
	dbHost    = "MONGO_HOST"
	dbHostKey = key(dbHost)
	dbPort    = "MONGO_PORT"
	dbPortKey = key(dbPort)
)

func configDB(ctx context.Context) (*mongo.Database, error) {

	ctx = context.WithValue(ctx, dbHostKey, os.Getenv(dbHost))
	ctx = context.WithValue(ctx, dbPortKey, os.Getenv(dbPort))

	// full uri with user and password
	// uri := fmt.Sprintf(`mongodb://%s:%s@%s:%s/%s`,
	// 	ctx.Value(usernameKey).(string),
	// 	ctx.Value(passwordKey).(string),
	// 	ctx.Value(dbHostKey).(string),
	// 	ctx.Value(dbPortKey).(string),
	// 	ctx.Value(databaseKey).(string),
	// )

	uri := fmt.Sprintf(`mongodb://%s:%s`,
		ctx.Value(dbHostKey).(string),
		ctx.Value(dbPortKey).(string),
	)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	log.Print("Connected to the DB!")
	return client.Database("test"), nil
}

func addUser(ctx context.Context, db *mongo.Database, usr User, collName string) (string, error) {

	password, err := bcrypt.GenerateFromPassword([]byte(usr.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("addUser: cannot create a password for the user: %v", err)
	}
	ver := Version{
		VerTag:  "1.0.1",
		Created: primitive.DateTime(time.Now().UnixNano() / int64(time.Millisecond)),
	}
	// usr.CreatedAt = primitive.DateTime(time.Now().UnixNano() / int64(time.Millisecond))

	usr.Password = string(password)
	usr.Version = ver

	res, err := db.Collection(collName).InsertOne(ctx, usr)

	if err != nil {
		return "", fmt.Errorf("addUser: task for to-do list couldn't be created: %v", err)
	}

	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}
