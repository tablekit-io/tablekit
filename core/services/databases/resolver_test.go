package databases_test

import (
	"context"
	"testing"

	"core/engine/identity"
	"core/services/databases"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDeriver returns a fixed identity per name and counts how often it is asked,
// so tests can prove the L1 cache avoids re-derivation.
type fakeDeriver struct {
	identities map[string]identity.Identity
	calls      map[string]int
}

func newFakeDeriver() *fakeDeriver {
	return &fakeDeriver{identities: map[string]identity.Identity{}, calls: map[string]int{}}
}

func (f *fakeDeriver) set(name string, id identity.Identity) { f.identities[name] = id }

func (f *fakeDeriver) DeriveIdentity(_ context.Context, name string) (identity.Identity, error) {
	f.calls[name]++
	return f.identities[name], nil
}

// fakeRepository is an in-memory Repository keyed by identity_key, mirroring the
// real upsert semantics (one stable id per physical database).
type fakeRepository struct {
	byKey map[string]databases.Record
	byID  map[uuid.UUID]databases.Record
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{byKey: map[string]databases.Record{}, byID: map[uuid.UUID]databases.Record{}}
}

func (f *fakeRepository) Upsert(_ context.Context, id uuid.UUID, record databases.Record) (uuid.UUID, error) {
	if existing, ok := f.byKey[record.IdentityKey]; ok {
		existing.Name = record.Name
		existing.Identity = record.Identity
		f.byKey[record.IdentityKey] = existing
		f.byID[existing.ID] = existing
		return existing.ID, nil
	}
	record.ID = id
	f.byKey[record.IdentityKey] = record
	f.byID[id] = record
	return id, nil
}

func (f *fakeRepository) Get(_ context.Context, id uuid.UUID) (*databases.Record, error) {
	record, ok := f.byID[id]
	if !ok {
		return nil, nil
	}
	return &record, nil
}

func TestResolveCachesAfterFirstDerive(t *testing.T) {
	deriver := newFakeDeriver()
	deriver.set("cafe", identity.Identity{Engine: "postgres", Key: "pg-cluster-1"})
	resolver := databases.NewResolver(deriver, newFakeRepository())
	ctx := context.Background()

	first, err := resolver.Resolve(ctx, "cafe")
	require.NoError(t, err)
	second, err := resolver.Resolve(ctx, "cafe")
	require.NoError(t, err)

	assert.Equal(t, first, second, "same name resolves to the same id")
	assert.Equal(t, 1, deriver.calls["cafe"], "the cache should avoid a second derivation")
}

func TestInvalidateCacheForcesRederive(t *testing.T) {
	deriver := newFakeDeriver()
	deriver.set("cafe", identity.Identity{Engine: "postgres", Key: "pg-cluster-1"})
	resolver := databases.NewResolver(deriver, newFakeRepository())
	ctx := context.Background()

	_, err := resolver.Resolve(ctx, "cafe")
	require.NoError(t, err)
	resolver.InvalidateCache()
	_, err = resolver.Resolve(ctx, "cafe")
	require.NoError(t, err)

	assert.Equal(t, 2, deriver.calls["cafe"], "invalidation should force a re-derivation")
}

func TestTwoNamesSamePhysicalDatabaseShareID(t *testing.T) {
	deriver := newFakeDeriver()
	same := identity.Identity{Engine: "postgres", Key: "pg-cluster-1"}
	deriver.set("cafe", same)
	deriver.set("cafe_alias", same)
	resolver := databases.NewResolver(deriver, newFakeRepository())
	ctx := context.Background()

	first, err := resolver.Resolve(ctx, "cafe")
	require.NoError(t, err)
	second, err := resolver.Resolve(ctx, "cafe_alias")
	require.NoError(t, err)

	assert.Equal(t, first, second, "two names for the same physical database collapse to one id")
}

func TestVerifyRefusesWhenNameRepointed(t *testing.T) {
	deriver := newFakeDeriver()
	deriver.set("cafe", identity.Identity{Engine: "postgres", Key: "pg-cluster-A"})
	resolver := databases.NewResolver(deriver, newFakeRepository())
	ctx := context.Background()

	// Save a query against physical database A.
	savedID, err := resolver.Resolve(ctx, "cafe")
	require.NoError(t, err)

	// The name is repointed at a different physical database (B); a reload clears
	// the cache so the next resolution re-derives.
	deriver.set("cafe", identity.Identity{Engine: "postgres", Key: "pg-cluster-B"})
	resolver.InvalidateCache()

	name, err := resolver.Verify(ctx, savedID)
	require.Error(t, err, "a repointed name must be refused")
	assert.Empty(t, name)
	assert.Contains(t, err.Error(), "different physical database")
}

func TestVerifyReturnsNameWhenIdentityUnchanged(t *testing.T) {
	deriver := newFakeDeriver()
	deriver.set("cafe", identity.Identity{Engine: "postgres", Key: "pg-cluster-A"})
	resolver := databases.NewResolver(deriver, newFakeRepository())
	ctx := context.Background()

	savedID, err := resolver.Resolve(ctx, "cafe")
	require.NoError(t, err)

	name, err := resolver.Verify(ctx, savedID)
	require.NoError(t, err)
	assert.Equal(t, "cafe", name)
}

func TestVerifyUnknownIDErrors(t *testing.T) {
	resolver := databases.NewResolver(newFakeDeriver(), newFakeRepository())

	unknown, err := uuid.NewV7()
	require.NoError(t, err)
	_, err = resolver.Verify(context.Background(), unknown)
	require.Error(t, err)
}
