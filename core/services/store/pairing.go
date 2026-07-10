package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"core/db/gen/tablekit/public/model"
	"core/db/gen/tablekit/public/table"

	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"

	. "github.com/go-jet/jet/v2/postgres"
)

// Pairing modes control whether a not-yet-paired client may pair.
const (
	PairingOnce       = "once"       // next new client pairs, then mode flips to disabled
	PairingIndefinite = "indefinite" // every new client may pair
	PairingDisabled   = "disabled"   // no new client may pair
)

// pairingModeKey is the config row that holds the current pairing mode.
const pairingModeKey = "pairing_mode"

// PairingRepository owns the paired state of clients (clients.paired_at) and the
// key/value config that holds the pairing mode. It bundles the two because
// TryPair reads the mode and writes both in one transaction.
type PairingRepository interface {
	TryPair(ctx context.Context, clientID uuid.UUID) (bool, error)
	SetPairingMode(ctx context.Context, mode string) error
	PairingStatus(ctx context.Context) (mode string, paired []uuid.UUID, err error)
}

type pairingRepository struct {
	database *sql.DB
	// mu serializes the TryPair read-modify-write so concurrent "once" pairings
	// can't both win (single-instance server).
	mu sync.Mutex
}

// NewPairingRepository returns a PairingRepository over the given database.
func NewPairingRepository(database *sql.DB) PairingRepository {
	return &pairingRepository{database: database}
}

// TryPair reports whether clientID may use the server, pairing it if the current
// mode allows. Already-paired clients are always allowed. The read-then-write is
// done in a transaction so concurrent authorizes cannot race the mode flip.
//
//   - disabled:   new clients rejected
//   - once:       new client paired, then mode flips to disabled
//   - indefinite: every new client paired; mode unchanged
func (r *pairingRepository) TryPair(ctx context.Context, clientID uuid.UUID) (bool, error) {
	// Serialize the read-check-write in-process so concurrent pairings under
	// "once" cannot both win a snapshot race (single-instance server).
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.database.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	var existing model.Clients
	err = SELECT(table.Clients.PairedAt).
		FROM(table.Clients).
		WHERE(table.Clients.ID.EQ(UUID(clientID))).
		QueryContext(ctx, tx, &existing)
	if errors.Is(err, qrm.ErrNoRows) {
		// The client must be registered before it can pair (authorize looks it up
		// first), so a missing row is unexpected rather than "not yet paired".
		return false, fmt.Errorf("pair unknown client %q", clientID)
	}
	if err != nil {
		return false, fmt.Errorf("check paired client: %w", err)
	}
	if existing.PairedAt != nil {
		return true, nil // already paired
	}

	mode, err := pairingMode(ctx, tx)
	if err != nil {
		return false, err
	}
	switch mode {
	case PairingIndefinite:
		if err := pairClient(ctx, tx, clientID); err != nil {
			return false, err
		}
	case PairingOnce:
		if err := pairClient(ctx, tx, clientID); err != nil {
			return false, err
		}
		if err := setPairingMode(ctx, tx, PairingDisabled); err != nil {
			return false, err
		}
	default: // disabled or unknown
		return false, nil
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return true, nil
}

// SetPairingMode persists a new pairing mode. Used by the `pairing` CLI.
func (r *pairingRepository) SetPairingMode(ctx context.Context, mode string) error {
	switch mode {
	case PairingOnce, PairingIndefinite, PairingDisabled:
	default:
		return fmt.Errorf("unknown pairing mode %q", mode)
	}
	return setPairingMode(ctx, r.database, mode)
}

// PairingStatus returns the current mode and the paired client ids.
func (r *pairingRepository) PairingStatus(ctx context.Context) (mode string, paired []uuid.UUID, err error) {
	mode, err = pairingMode(ctx, r.database)
	if err != nil {
		return "", nil, err
	}
	var rows []model.Clients
	err = SELECT(table.Clients.ID).
		FROM(table.Clients).
		WHERE(table.Clients.PairedAt.IS_NOT_NULL()).
		ORDER_BY(table.Clients.ID.ASC()).
		QueryContext(ctx, r.database, &rows)
	if err != nil {
		return "", nil, fmt.Errorf("list paired clients: %w", err)
	}
	paired = make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		paired = append(paired, row.ID)
	}
	return mode, paired, nil
}

// ---- config + pairing helpers (work against either *sql.DB or *sql.Tx) --

// getConfig reads the JSON value under key into dest, reporting whether the row
// existed. A missing key is not an error: it returns (false, nil) so callers can
// apply their own default.
func getConfig(ctx context.Context, q qrm.Queryable, key string, dest any) (bool, error) {
	var row model.Config
	err := SELECT(table.Config.Value).
		FROM(table.Config).
		WHERE(table.Config.Key.EQ(String(key))).
		QueryContext(ctx, q, &row)
	if errors.Is(err, qrm.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read config %q: %w", key, err)
	}
	if err := json.Unmarshal(row.Value, dest); err != nil {
		return false, fmt.Errorf("decode config %q: %w", key, err)
	}
	return true, nil
}

// setConfig upserts value (JSON-encoded) under key. The value is sent as text and
// cast into the jsonb column by Postgres, matching the read path.
func setConfig(ctx context.Context, q qrm.Executable, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode config %q: %w", key, err)
	}
	stmt := table.Config.
		INSERT(table.Config.Key, table.Config.Value).
		VALUES(key, string(raw)).
		ON_CONFLICT(table.Config.Key).
		DO_UPDATE(SET(table.Config.Value.SET(table.Config.EXCLUDED.Value)))
	if _, err := stmt.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("set config %q: %w", key, err)
	}
	return nil
}

// pairingMode reads the current mode, defaulting to PairingOnce when the config
// row is absent (matching the old JSON default).
func pairingMode(ctx context.Context, q qrm.Queryable) (string, error) {
	var mode string
	found, err := getConfig(ctx, q, pairingModeKey, &mode)
	if err != nil {
		return "", err
	}
	if !found {
		return PairingOnce, nil
	}
	return mode, nil
}

// setPairingMode upserts the pairing_mode config value.
func setPairingMode(ctx context.Context, q qrm.Executable, mode string) error {
	return setConfig(ctx, q, pairingModeKey, mode)
}

// pairClient marks clientID paired by stamping clients.paired_at with the current
// time. Callers only reach here for a client that is not yet paired.
func pairClient(ctx context.Context, q qrm.Executable, clientID uuid.UUID) error {
	stmt := table.Clients.
		UPDATE(table.Clients.PairedAt).
		SET(TimestampzT(time.Now())).
		WHERE(table.Clients.ID.EQ(UUID(clientID)))
	if _, err := stmt.ExecContext(ctx, q); err != nil {
		return fmt.Errorf("pair client: %w", err)
	}
	return nil
}
