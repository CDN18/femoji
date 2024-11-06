package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "0.5.0"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Femoji",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Femoji CLI v%s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
