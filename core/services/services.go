// Package services bundles the long-lived, process-wide dependencies — config,
// persistence (store), the read-only query engine, and the JWT issuer — into one
// value constructed at startup and threaded through the app. It is the single
// place those services are wired, and the extension point for any future ones.
package services

import (
	"database/sql"
	"time"

	"core/db"
	"core/engine"
	"core/services/config"
	"core/services/oauth"
	"core/services/queries"
	"core/services/store"
)

// Services holds the shared services every layer needs.
type Services struct {
	Config  *config.Config
	Store   *store.Store
	Engine  *engine.Service
	Issuer  *oauth.Issuer
	DB      *sql.DB
	Queries *queries.Repository
}

// New loads config from the environment, opens the on-disk store, and loads the
// configured databases.
func New() (*Services, error) {
	configService := config.Load()
	// Open the Postgres database first: the store persists all of its OAuth state
	// in it (the signing key lives only in the SIGNING_KEY env, not on disk).
	database, err := db.Open(configService.DatabaseDSN())
	if err != nil {
		return nil, err
	}
	storageService := store.New(database)
	// Resolve .yaml/.yml by base name (dies if both exist), then remember the
	// resolved path so the app reports what it actually loaded.
	resolvedDatabasesFile, err := config.ResolveDatabasesFile(configService.DatabasesFile)
	if err != nil {
		return nil, err
	}
	configService.DatabasesFile = resolvedDatabasesFile
	engineService, err := engine.Load(resolvedDatabasesFile, engine.Limits{
		StatementTimeout: 10 * time.Second,
		MaxRows:          2048,
		MaxBytes:         64 * 1024,
	})
	if err != nil {
		return nil, err
	}
	issuer, err := oauth.NewIssuer(configService)
	if err != nil {
		return nil, err
	}
	return &Services{
		Config:  configService,
		Store:   storageService,
		Engine:  engineService,
		Issuer:  issuer,
		DB:      database,
		Queries: queries.New(database),
	}, nil
}

// Close releases resources held by the services (currently the database handle).
// Call it on shutdown.
func (s *Services) Close() error {
	if s.DB != nil {
		return s.DB.Close()
	}
	return nil
}
