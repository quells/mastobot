package cmd

import (
	"github.com/quells/mastobot/internal/oauth2"
	"github.com/spf13/cobra"
)

var (
	appName   string
	userEmail string
	password  string
)

func init() {
	appCmd.PersistentFlags().StringVar(&appName, "name", "", "Name of the application")
	must(appCmd.MarkPersistentFlagRequired("name"))

	appTokenRenewCmd.Flags().StringVarP(&userEmail, "email", "U", "", "Account email")
	appTokenRenewCmd.Flags().StringVarP(&password, "password", "P", "", "Account password")
	must(appTokenRenewCmd.MarkFlagRequired("email"))
	must(appTokenRenewCmd.MarkFlagRequired("password"))
	appTokenCmd.AddCommand(appTokenRenewCmd)

	appTokenCmd.AddCommand(appTokenRevokeCmd)

	appCmd.AddCommand(appRegisterCmd)
	appCmd.AddCommand(appTokenCmd)
	appCmd.AddCommand(appTootCmd)

	rootCmd.AddCommand(appCmd)
}

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Application Helpers",
	Long: `Application Helpers
Register an application with an instance.
Generate OAuth2 access tokens.`,
}

var appRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register Application",
	Long:  "Register an application with an instance.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return oauth2.RegisterApp(cmd.Context(), instance, appName)
	},
}

var appTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Access Token Helpers",
	Long: `Access Token Helpers to connect to an instance as a user.
Renew an access token.
Revoke an access token.`,
}

var appTokenRenewCmd = &cobra.Command{
	Use:   "renew",
	Short: "Renew an access token",
	RunE: func(cmd *cobra.Command, args []string) error {
		return oauth2.GetAccessToken(cmd.Context(), instance, appName, userEmail, password)
	},
}

var appTokenRevokeCmd = &cobra.Command{}

var appTootCmd = &cobra.Command{
	Use:   "toot",
	Short: "Toot!",
}
