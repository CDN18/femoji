package cmd

import (
	"github.com/spf13/cobra"

	"github.com/CDN18/femoji-cli/internal/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Log in or out",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in",
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.Login(User)
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out",
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.Logout(User)
	},
}

var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display the default currently authenticated user, if there is one",
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.Whoami()
	},
}

func init() {
	rootCmd.AddCommand(authCmd)

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authWhoamiCmd)
}
