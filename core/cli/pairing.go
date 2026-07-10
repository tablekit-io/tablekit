package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"core/services"
	"core/services/oauth"
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
		return applyPairingMode(cmd.Context(), mode)
	},
}

var pairingDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Reject any new client from pairing",
	RunE: func(cmd *cobra.Command, args []string) error {
		return applyPairingMode(cmd.Context(), store.PairingDisabled)
	},
}

var pairingTokenGenerateCmd = &cobra.Command{
	Use:   "token:generate",
	Short: "Mint a long-lived static token (valid 6 months) for MCP access",
	Long: "Mint a long-lived static token a client can present as\n" +
		"`Authorization: Bearer <token>` to reach /mcp without the OAuth flow.\n" +
		"Note the printed token id: it is needed to revoke the token later.",
	RunE: func(cmd *cobra.Command, args []string) error {
		appServices, err := services.New()
		if err != nil {
			return err
		}

		minted, err := oauth.MintStatic(cmd.Context(), appServices.Clients, appServices.StaticTokens, appServices.Issuer)
		if err != nil {
			return err
		}

		fmt.Printf("token id: %s\n", minted.ID)
		fmt.Printf("expires:  %s\n", minted.ExpiresAt.Format(time.RFC3339))
		// Print the bearer value on its own final line so scripts can capture it.
		fmt.Println(minted.Token)
		return nil
	},
}

var pairingTokenRevokeCmd = &cobra.Command{
	Use:   "token:revoke <id OR token>",
	Short: "Revoke a static token by its id or the token itself",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appServices, err := services.New()
		if err != nil {
			return err
		}

		// Accept either the static token id or the token itself, so a user who
		// kept the token but lost the id can still revoke it.
		rawID := args[0]
		if rawJWT, ok := strings.CutPrefix(rawID, oauth.TokenPrefix); ok {
			id, err := oauth.StaticTokenID(rawJWT)
			if err != nil {
				return fmt.Errorf("could not read token id from token: %w", err)
			}
			rawID = id
		}

		tokenID, err := uuid.Parse(rawID)
		if err != nil {
			return fmt.Errorf("invalid static token id %q", rawID)
		}
		if err := appServices.StaticTokens.RevokeStaticToken(cmd.Context(), tokenID); err != nil {
			return err
		}
		fmt.Printf("revoked static token %s\n", tokenID)
		return nil
	},
}

// applyPairingMode persists the mode and prints the result. It writes to the same
// state database the server reads, so a running server picks up the change on its
// next authorize.
func applyPairingMode(ctx context.Context, mode string) error {
	appServices, err := services.New()
	if err != nil {
		return err
	}
	if err := appServices.Pairing.SetPairingMode(ctx, mode); err != nil {
		return err
	}
	_, paired, err := appServices.Pairing.PairingStatus(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("pairing mode: %s (%d client(s) paired)\n", mode, len(paired))
	return nil
}

func init() {
	pairingEnableCmd.Flags().BoolVar(&pairingOnce, "once", false, "admit a single next client, then lock")
	pairingEnableCmd.Flags().BoolVar(&pairingIndefinitely, "indefinitely", false, "keep pairing open for any number of clients")

	pairingCmd.AddCommand(pairingEnableCmd, pairingDisableCmd, pairingTokenGenerateCmd, pairingTokenRevokeCmd)
}
