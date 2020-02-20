package models

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

// GetDB returns *mongo.Database instance
func GetDB() *mongo.Database {
	return db
}

// InitDB initializes DB connections
func InitDB(URI, databaseName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(URI))
	if err != nil {
		return fmt.Errorf("on connecting to database endpoint: %s", err.Error())
	}

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err = client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("on pinging to database endpoint: %s", err.Error())
	}

	db = client.Database(databaseName)
	return nil
}
