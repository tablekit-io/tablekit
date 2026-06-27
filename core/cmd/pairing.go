package cmd

import (
	"fmt"

	"core/services"
	"core/store"

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
	pairingCmd.AddCommand(pairingEnableCmd, pairingDisableCmd)
	rootCmd.AddCommand(pairingCmd)
}
