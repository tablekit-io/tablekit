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
	"core/services/databases"
	"core/services/oauth"
	"core/services/queries"
	"core/services/requests"
	"core/services/store"
)

// Services holds the shared services every layer needs. The store aggregate is
// exposed as one repository per concern rather than a single handle.
type Services struct {
	Config       *config.Config
	Clients      store.ClientRepository
	AuthCodes    store.AuthCodeRepository
	TokenChains  store.TokenChainRepository
	StaticTokens store.StaticTokenRepository
	Pairing      store.PairingRepository
	Engine       *engine.Service
	Issuer       *oauth.Issuer
	DB           *sql.DB
	Queries      queries.QueryRepository
	Requests     requests.RequestLog
	Databases    *databases.Resolver
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
	// Resolve .yaml/.yml by base name (dies if both exist) for the initial load.
	// Config.DatabasesFile stays the originally configured path so the file
	// watcher can re-resolve it (e.g. a .yaml appearing after a .yml, or a file
	// created after startup) — see WatchDatabasesFile.
	resolvedDatabasesFile, err := config.ResolveDatabasesFile(configService.DatabasesFile)
	if err != nil {
		return nil, err
	}
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
		Config:       configService,
		Clients:      store.NewClientRepository(database),
		AuthCodes:    store.NewAuthCodeRepository(database),
		TokenChains:  store.NewTokenChainRepository(database),
		StaticTokens: store.NewStaticTokenRepository(database),
		Pairing:      store.NewPairingRepository(database),
		Engine:       engineService,
		Issuer:       issuer,
		DB:           database,
		Queries:      queries.New(database),
		Requests:     requests.New(database),
		Databases:    databases.NewResolver(engineService, databases.NewRepository(database)),
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
