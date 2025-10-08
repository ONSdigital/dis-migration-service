package api

import (
	"context"
	"github.com/ONSdigital/dis-migration-service/store"

	"github.com/gorilla/mux"
)

// API provides a struct to wrap the api around
type API struct {
	Router *mux.Router
	Store  *store.Datastore
}

// Setup function sets up the api and returns an api
func Setup(ctx context.Context, r *mux.Router, dataStore *store.Datastore) *API {
	api := &API{
		Router: r,
		Store:  dataStore,
	}

	// TODO: remove hello world example handler route
	r.HandleFunc("/hello", HelloHandler(ctx)).Methods("GET")
	return api
}
