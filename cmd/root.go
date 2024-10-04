package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "femoji",
	Short: "Femoji is a tool for managing custom emojis on Fediverse instances",
}

var User string
var File string

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&User, "user", "u", "", "username@domain of the account whose data we're working with")
}
