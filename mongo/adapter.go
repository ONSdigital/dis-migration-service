package mongo

import (
	"context"

	mongodriver "github.com/ONSdigital/dp-mongodb/v3/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

// mongoConnectionAdapter wraps a real MongoConnection to implement our
// interface
type mongoConnectionAdapter struct {
	conn *mongodriver.MongoConnection
}

// NewMongoConnectionAdapter creates an adapter that wraps a real MongoDB
// connection
func NewMongoConnectionAdapter(conn *mongodriver.MongoConnection) MongoConnection {
	return &mongoConnectionAdapter{conn: conn}
}

// Collection returns a wrapped collection that implements our
// MongoCollection interface
func (a *mongoConnectionAdapter) Collection(name string) MongoCollection {
	return &mongoCollectionAdapter{coll: a.conn.Collection(name)}
}

// RunCommand executes a database command
func (a *mongoConnectionAdapter) RunCommand(ctx context.Context, command interface{}) error {
	return a.conn.RunCommand(ctx, command)
}

// ListCollectionsFor lists all collections in a database
func (a *mongoConnectionAdapter) ListCollectionsFor(ctx context.Context, database string) ([]string, error) {
	return a.conn.ListCollectionsFor(ctx, database)
}

// mongoCollectionAdapter wraps a real MongoDB collection to implement
// our interface
type mongoCollectionAdapter struct {
	coll *mongodriver.Collection
}

// InsertOne wraps the real InsertOne method
func (c *mongoCollectionAdapter) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	result, err := c.coll.InsertOne(ctx, document)
	if err != nil {
		return nil, err
	}
	// Convert dp-mongodb result to standard mongo driver result
	return &mongo.InsertOneResult{InsertedID: result.InsertedId}, nil
}

// FindOne wraps the real FindOne method
func (c *mongoCollectionAdapter) FindOne(ctx context.Context, filter, result interface{}, opts ...interface{}) error {
	// Convert opts to dp-mongodb options
	dpOpts := make([]mongodriver.FindOption, 0, len(opts))
	for _, opt := range opts {
		if dpOpt, ok := opt.(mongodriver.FindOption); ok {
			dpOpts = append(dpOpts, dpOpt)
		}
	}
	return c.coll.FindOne(ctx, filter, result, dpOpts...)
}

// Find wraps the real Find method
func (c *mongoCollectionAdapter) Find(ctx context.Context, filter, results interface{}, opts ...interface{}) (int, error) {
	// Convert opts to dp-mongodb options
	dpOpts := make([]mongodriver.FindOption, 0, len(opts))
	for _, opt := range opts {
		if dpOpt, ok := opt.(mongodriver.FindOption); ok {
			dpOpts = append(dpOpts, dpOpt)
		}
	}
	return c.coll.Find(ctx, filter, results, dpOpts...)
}

// UpdateOne wraps the real UpdateOne method
func (c *mongoCollectionAdapter) UpdateOne(ctx context.Context, filter, update interface{}, opts ...interface{}) (*mongo.UpdateResult, error) {
	// dp-mongodb UpdateOne doesn't take variadic options
	result, err := c.coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	// Convert dp-mongodb result to standard mongo driver result
	return &mongo.UpdateResult{
		MatchedCount:  int64(result.MatchedCount),
		ModifiedCount: int64(result.ModifiedCount),
		UpsertedCount: int64(result.UpsertedCount),
		UpsertedID:    result.UpsertedID,
	}, nil
}

// FindOneAndUpdate wraps the real FindOneAndUpdate method
func (c *mongoCollectionAdapter) FindOneAndUpdate(ctx context.Context, filter, update, result interface{}, opts ...interface{}) error {
	// Convert opts to dp-mongodb options
	dpOpts := make([]mongodriver.FindOption, 0, len(opts))
	for _, opt := range opts {
		if dpOpt, ok := opt.(mongodriver.FindOption); ok {
			dpOpts = append(dpOpts, dpOpt)
		}
	}
	return c.coll.FindOneAndUpdate(ctx, filter, update, result, dpOpts...)
}
