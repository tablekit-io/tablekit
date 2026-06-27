package cli

import (
	"fmt"
	"strings"
	"time"

	"core/http/app/oauth"
	"core/services"
	"core/services/store"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var pairingCmd = &cobra.Command{
	Use:   "pairing",
	Short: "Manage which clients may pair with the server",
}

var (
	pairingOnce         bool
	pairingIndefinitely bool
)

var pairingEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Allow client(s) to pair",
	Long: "Allow client(s) to pair. Use --once to admit a single next client, or\n" +
		"--indefinitely to keep pairing open for any number of clients.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if pairingOnce == pairingIndefinitely {
			return fmt.Errorf("specify exactly one of --once or --indefinitely")
		}
		mode := store.PairingOnce
		if pairingIndefinitely {
			mode = store.PairingIndefinite
		}
		return applyPairingMode(mode)
	},
}

var pairingDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Reject any new client from pairing",
	RunE: func(cmd *cobra.Command, args []string) error {
		return applyPairingMode(store.PairingDisabled)
	},
}

var pairingTokenGenerateCmd = &cobra.Command{
	Use:   "token:generate",
	Short: "Mint a long-lived bearer token (valid 6 months) for MCP access",
	Long: "Mint a long-lived bearer token a client can present as\n" +
		"`Authorization: Bearer <token>` to reach /mcp without the OAuth flow.\n" +
		"Note the printed token id: it is needed to revoke the token later.",
	RunE: func(cmd *cobra.Command, args []string) error {
		appServices, err := services.New()
		if err != nil {
			return err
		}
		issuer, err := oauth.NewIssuer(appServices)
		if err != nil {
			return err
		}

		clientID := uuid.NewString()
		tokenID := uuid.NewString()
		now := time.Now()

		// A bearer token registers as its own client: no redirect URIs, no name.
		if err := appServices.Store.SaveClient(&store.Client{
			ClientID:     clientID,
			ClientName:   nil,
			RedirectURIs: []string{},
			Type:         "bearer",
			CreatedAt:    now,
		}); err != nil {
			return err
		}

		token, expiresAt, err := issuer.IssueBearer(clientID, tokenID)
		if err != nil {
			return err
		}
		if err := appServices.Store.PutBearerToken(&store.BearerToken{
			ID:        tokenID,
			ClientID:  clientID,
			CreatedAt: now,
			ExpiresAt: expiresAt,
		}); err != nil {
			return err
		}

		fmt.Printf("token id: %s\n", tokenID)
		fmt.Printf("expires:  %s\n", expiresAt.Format(time.RFC3339))
		// Print the bearer value on its own final line so scripts can capture it.
		fmt.Println(oauth.TokenPrefix + token)
		return nil
	},
}

var pairingTokenRevokeCmd = &cobra.Command{
	Use:   "token:revoke <id OR token>",
	Short: "Revoke a bearer token by its id or the token itself",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appServices, err := services.New()
		if err != nil {
			return err
		}

		// Accept either the bearer token id or the token itself, so a user who
		// kept the token but lost the id can still revoke it.
		tokenID := args[0]
		if rawJWT, ok := strings.CutPrefix(tokenID, oauth.TokenPrefix); ok {
			id, err := oauth.BearerTokenID(rawJWT)
			if err != nil {
				return fmt.Errorf("could not read token id from token: %w", err)
			}
			tokenID = id
		}

		if err := appServices.Store.RevokeBearerToken(tokenID); err != nil {
			return err
		}
		fmt.Printf("revoked bearer token %s [data dir: %s]\n", tokenID, appServices.Config.DataDir)
		return nil
	},
}

// applyPairingMode persists the mode to the data dir and prints the result.
// It targets the same DATA_DIR the server reads, so a running server picks up
// the change on its next authorize.
func applyPairingMode(mode string) error {
	appServices, err := services.New()
	if err != nil {
		return err
	}
	if err := appServices.Store.SetPairingMode(mode); err != nil {
		return err
	}
	_, paired, err := appServices.Store.PairingStatus()
	if err != nil {
		return err
	}
	fmt.Printf("pairing mode: %s (%d client(s) paired) [data dir: %s]\n", mode, len(paired), appServices.Config.DataDir)
	return nil
}

func init() {
	pairingEnableCmd.Flags().BoolVar(&pairingOnce, "once", false, "admit a single next client, then lock")
	pairingEnableCmd.Flags().BoolVar(&pairingIndefinitely, "indefinitely", false, "keep pairing open for any number of clients")
	pairingCmd.AddCommand(pairingEnableCmd, pairingDisableCmd, pairingTokenGenerateCmd, pairingTokenRevokeCmd)
	rootCmd.AddCommand(pairingCmd)
}
