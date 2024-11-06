package cmd

import (
	"github.com/spf13/cobra"

	"github.com/CDN18/femoji-cli/internal/auth"
	"github.com/CDN18/femoji-cli/internal/download"
)

var instanceType string

var downloadCmd = &cobra.Command{
	Use:   "download [instance] [category] --software mastodon|misskey",
	Short: "Download emojis from an instance",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		authClient, err := auth.NewAuthClient(User)
		if err != nil {
			return err
		}

		instance := "DEFAULT"
		if len(args) > 0 {
			instance = args[0]
		}
		category := "*"
		if len(args) > 1 {
			category = args[1]
		}

		return download.Download(authClient, instance, category, override, instanceType)
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().BoolVar(&override, "override", false, "Override existing files when downloading")
	downloadCmd.Flags().StringVar(&instanceType, "software", "mastodon", "Instance type (mastodon or misskey)")
}
