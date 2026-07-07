package databases

import (
	"context"
	"fmt"
	"sync"

	"core/engine/identity"

	"github.com/google/uuid"
)

// identityDeriver connects to a configured database and fingerprints the physical
// database behind it. *engine.Service satisfies this; taking the narrow interface
// (rather than the concrete engine) keeps the resolver unit-testable and avoids
// depending on the whole engine surface.
type identityDeriver interface {
	DeriveIdentity(ctx context.Context, databaseName string) (identity.Identity, error)
}

// Resolver maps a configured database name to a stable database_id, deriving and
// persisting the physical-database identity on first use. It holds an in-memory
// L1 cache (name -> database_id) so only the first query for a name pays the
// derivation round-trip; the cache is cleared whenever databases.yaml reloads,
// since that is exactly when a name's connection details can change.
type Resolver struct {
	deriver identityDeriver
	repo    Repository

	mu    sync.RWMutex
	cache map[string]uuid.UUID
}

// NewResolver wires a Resolver to the identity deriver (the engine) and the
// databases repository.
func NewResolver(deriver identityDeriver, repo Repository) *Resolver {
	return &Resolver{
		deriver: deriver,
		repo:    repo,
		cache:   make(map[string]uuid.UUID),
	}
}

// Resolve returns the stable database_id for a configured database name,
// deriving and persisting its physical identity on first use. This is the lazy
// first-query step: a cache hit returns immediately; a miss connects via the
// driver, upserts the identity (minting a new id or matching an existing
// physical database), caches the result, and returns it.
func (r *Resolver) Resolve(ctx context.Context, name string) (uuid.UUID, error) {
	if id, ok := r.cacheGet(name); ok {
		return id, nil
	}

	derived, err := r.deriver.DeriveIdentity(ctx, name)
	if err != nil {
		return uuid.Nil, err
	}

	newID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate database id: %w", err)
	}

	id, err := r.repo.Upsert(ctx, newID, Record{
		Name:        name,
		Type:        derived.Engine,
		IdentityKey: derived.Key,
		Identity:    derived.Attributes,
	})
	if err != nil {
		return uuid.Nil, err
	}

	r.cachePut(name, id)
	return id, nil
}

// Verify recovers the configured name a stored query was saved against, re-derives
// its current physical identity, and confirms the name still points at the SAME
// physical database. It returns that name (to run the query against) or an error
// when the name now resolves to a different physical database — the guard that
// stops a stored query from silently running against a repointed database.
func (r *Resolver) Verify(ctx context.Context, databaseID uuid.UUID) (string, error) {
	record, err := r.repo.Get(ctx, databaseID)
	if err != nil {
		return "", err
	}
	if record == nil {
		return "", fmt.Errorf("the database for this query is no longer known")
	}

	currentID, err := r.Resolve(ctx, record.Name)
	if err != nil {
		return "", fmt.Errorf("re-check database %q: %w", record.Name, err)
	}
	if currentID != databaseID {
		return "", fmt.Errorf("database %q now points to a different physical database than when this query was saved", record.Name)
	}
	return record.Name, nil
}

// InvalidateCache clears the L1 cache. It is called on every databases.yaml
// reload so the next query re-derives instead of trusting a stale name->id
// mapping (a reload is the moment a name's connection details can change).
func (r *Resolver) InvalidateCache() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = make(map[string]uuid.UUID)
}

func (r *Resolver) cacheGet(name string) (uuid.UUID, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	id, ok := r.cache[name]
	return id, ok
}

func (r *Resolver) cachePut(name string, id uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[name] = id
}
