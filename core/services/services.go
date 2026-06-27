// Package services bundles the long-lived, process-wide dependencies (config
// and persistence) into one value constructed at startup and threaded through
// the app. It is the single place those services are wired, and the extension
// point for any future ones.
package services

import (
	"core/config"
	"core/store"
)

// Services holds the shared services every layer needs.
type Services struct {
	Config *config.Config
	Store  *store.Store
}

// New loads config from the environment and opens the on-disk store.
func New() (*Services, error) {
	configService := config.Load()
	storageService, err := store.New(configService.DataDir)
	if err != nil {
		return nil, err
	}
	return &Services{Config: configService, Store: storageService}, nil
}
