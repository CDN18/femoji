package cmd

import (
	"github.com/spf13/cobra"

	"github.com/CDN18/femoji-cli/internal/auth"
	"github.com/CDN18/femoji-cli/internal/download"
)

var downloadCmd = &cobra.Command{
	Use:   "download [instance]",
	Short: "Download emojis from an instance",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		authClient, err := auth.NewAuthClient(User)
		if err != nil {
			return err
		}

		instance := "DEFAULT"
		if len(args) > 0 {
			instance = args[0]
		}

		return download.Download(authClient, instance, override)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().BoolVar(&override, "override", false, "Override existing files when downloading")
}
