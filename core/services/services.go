// Package services bundles the long-lived, process-wide dependencies — config,
// persistence (store), the read-only query engine, and the JWT issuer — into one
// value constructed at startup and threaded through the app. It is the single
// place those services are wired, and the extension point for any future ones.
package services

import (
	"time"

	"core/engine"
	"core/services/config"
	"core/services/oauth"
	"core/services/store"
)

// Services holds the shared services every layer needs.
type Services struct {
	Config *config.Config
	Store  *store.Store
	Engine *engine.Service
	Issuer *oauth.Issuer
}

// New loads config from the environment, opens the on-disk store, and loads the
// configured databases.
func New() (*Services, error) {
	configService := config.Load()
	storageService, err := store.New(configService.DataDir)
	if err != nil {
		return nil, err
	}
	engineService, err := engine.Load(configService.DatabasesFile, engine.Limits{
		StatementTimeout: 10 * time.Second,
		MaxRows:          2048,
		MaxBytes:         64 * 1024,
	})
	if err != nil {
		return nil, err
	}
	issuer, err := oauth.NewIssuer(configService, storageService)
	if err != nil {
		return nil, err
	}
	return &Services{
		Config: configService,
		Store:  storageService,
		Engine: engineService,
		Issuer: issuer,
	}, nil
}
