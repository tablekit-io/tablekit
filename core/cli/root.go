package cli

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tablekit",
	Short: "tablekit core service",
}

// Execute runs the root command and exits on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("command failed")
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(pairingCmd)
}
