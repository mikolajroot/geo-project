package database

import (
	"context"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	clientInstance *mongo.Client
	clientError    error
	mongoOnce      sync.Once
)

func GetMongoClient(uri string) (*mongo.Client, error) {
	mongoOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI(uri)
		client, err := mongo.Connect(clientOptions)
		if err != nil {
			clientError = err
			return
		}

		if err = client.Ping(ctx, nil); err != nil {
			clientError = err
			return
		}

		clientInstance = client
	})

	return clientInstance, clientError
}