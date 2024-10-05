package cmd

import (
	"github.com/spf13/cobra"

	"github.com/CDN18/femoji-cli/internal/auth"
	"github.com/CDN18/femoji-cli/internal/upload"
)

var (
	override bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload <path> <category>",
	Short: "Upload emojis from a directory",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		authClient, err := auth.NewAuthClient(User)
		if err != nil {
			return err
		}
		return upload.Upload(authClient, args[0], args[1], override)
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().BoolVar(&override, "override", false, "Override existing emojis with the same shortcode")
}
