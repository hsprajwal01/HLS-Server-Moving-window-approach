package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "app",
	Short: "HLS Streaming Server",
	Long:  `A simple tool to manage HLS streaming, dynamic ad insertion, and manifest updates.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
