package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

//go:generate moq -out mock/mongo_collection.go -pkg mock . MongoCollection
//go:generate moq -out mock/mongo_connection.go -pkg mock . MongoConnection

// MongoCollection defines the interface for MongoDB collection operations
type MongoCollection interface {
	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)
	FindOne(ctx context.Context, filter interface{}, result interface{}, opts ...interface{}) error
	Find(ctx context.Context, filter interface{}, results interface{}, opts ...interface{}) (int, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*mongo.UpdateResult, error)
	FindOneAndUpdate(ctx context.Context, filter interface{}, update interface{}, result interface{}, opts ...interface{}) error
}

// MongoConnection defines the interface for MongoDB connection operations
type MongoConnection interface {
	Collection(name string) MongoCollection
	RunCommand(ctx context.Context, command interface{}) error
	ListCollectionsFor(ctx context.Context, database string) ([]string, error)
}
