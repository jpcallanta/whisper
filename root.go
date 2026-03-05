package main

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "whisper",
	Short: "Migrate AWS Secrets Manager secrets between regions",
}

// Registers subcommands with the root command.
func init() {
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(dumpCmd)
}

// Runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
