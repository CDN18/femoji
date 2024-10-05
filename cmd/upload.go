package cmd

import (
	"github.com/spf13/cobra"

	"github.com/CDN18/femoji-cli/internal/auth"
	"github.com/CDN18/femoji-cli/internal/upload"
)

var uploadCmd = &cobra.Command{
	Use:   "upload <path> [category]",
	Short: "Upload emojis from a directory",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		authClient, err := auth.NewAuthClient(User)
		if err != nil {
			return err
		}
		path := args[0]
		category := "uncategorized"
		if len(args) == 2 {
			category = args[1]
		}
		return upload.Upload(authClient, path, category, override)
	},
}

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().BoolVar(&override, "override", false, "Override existing emojis with the same shortcode")
}
