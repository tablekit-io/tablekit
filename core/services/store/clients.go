package store

import (
	"fmt"
	"slices"
	"time"
)

// Client is a registered OAuth client (RFC 7591 dynamic registration). A CLI-
// minted bearer token also registers a Client, with Type "bearer", a nil
// ClientName (serialized as null) and an empty RedirectURIs.
type Client struct {
	ClientID string `json:"client_id"`
	// ClientName is a pointer so an absent name serializes as JSON null rather
	// than being omitted, which is what bearer clients carry.
	ClientName   *string  `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
	// Type distinguishes a CLI bearer client ("bearer") from an OAuth client
	// (empty/omitted).
	Type      string    `json:"type,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Pairing modes control whether a not-yet-paired client may pair.
const (
	PairingOnce       = "once"       // next new client pairs, then mode flips to disabled
	PairingIndefinite = "indefinite" // every new client may pair
	PairingDisabled   = "disabled"   // no new client may pair
)

// clientsFile is the on-disk shape of clients.json.
type clientsFile struct {
	// PairingMode gates whether new clients may pair; defaults to "once".
	PairingMode string `json:"pairing_mode"`
	// PairedClientIDs are the clients allowed to authenticate.
	PairedClientIDs []string `json:"paired_client_ids"`
	// LegacyPairedClientID migrates the old single-client field: it is folded
	// into PairedClientIDs on load and dropped on the next save (omitempty).
	LegacyPairedClientID string             `json:"paired_client_id,omitempty"`
	Clients              map[string]*Client `json:"clients"`
}

func (s *Store) loadClients() (*clientsFile, error) {
	clientsData := &clientsFile{Clients: map[string]*Client{}}
	if err := s.readJSON("clients.json", clientsData); err != nil {
		return nil, err
	}
	if clientsData.Clients == nil {
		clientsData.Clients = map[string]*Client{}
	}
	if clientsData.PairedClientIDs == nil {
		// Keep it a non-nil slice so it marshals to [] not null (schema: array).
		clientsData.PairedClientIDs = []string{}
	}
	if clientsData.PairingMode == "" {
		clientsData.PairingMode = PairingOnce
	}
	// Fold the legacy single paired client into the list.
	if clientsData.LegacyPairedClientID != "" {
		if !slices.Contains(clientsData.PairedClientIDs, clientsData.LegacyPairedClientID) {
			clientsData.PairedClientIDs = append(clientsData.PairedClientIDs, clientsData.LegacyPairedClientID)
		}
		clientsData.LegacyPairedClientID = ""
	}
	return clientsData, nil
}

// SaveClient persists a newly registered client.
func (s *Store) SaveClient(c *Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return err
	}
	clientsData.Clients[c.ClientID] = c
	return s.writeJSON("clients.json", clientsData)
}

// GetClient returns the client by id, or nil if unknown.
func (s *Store) GetClient(id string) (*Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return nil, err
	}
	return clientsData.Clients[id], nil
}

// TryPair reports whether clientID may use the server, pairing it if the
// current mode allows. Already-paired clients are always allowed.
//
//   - disabled:   new clients rejected
//   - once:       new client paired, then mode flips to disabled
//   - indefinite: every new client paired; mode unchanged
func (s *Store) TryPair(clientID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return false, err
	}
	if slices.Contains(clientsData.PairedClientIDs, clientID) {
		return true, nil
	}
	switch clientsData.PairingMode {
	case PairingIndefinite:
		clientsData.PairedClientIDs = append(clientsData.PairedClientIDs, clientID)
	case PairingOnce:
		clientsData.PairedClientIDs = append(clientsData.PairedClientIDs, clientID)
		clientsData.PairingMode = PairingDisabled
	default: // disabled or unknown
		return false, nil
	}
	if err := s.writeJSON("clients.json", clientsData); err != nil {
		return false, err
	}
	return true, nil
}

// SetPairingMode persists a new pairing mode. Used by the `pairing` CLI.
func (s *Store) SetPairingMode(mode string) error {
	switch mode {
	case PairingOnce, PairingIndefinite, PairingDisabled:
	default:
		return fmt.Errorf("unknown pairing mode %q", mode)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return err
	}
	clientsData.PairingMode = mode
	return s.writeJSON("clients.json", clientsData)
}

// PairingStatus returns the current mode and the paired client ids.
func (s *Store) PairingStatus() (mode string, paired []string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	clientsData, err := s.loadClients()
	if err != nil {
		return "", nil, err
	}
	return clientsData.PairingMode, clientsData.PairedClientIDs, nil
}
