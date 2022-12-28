package cmd

import (
	"fmt"
	"github.com/quells/mastobot/internal/oauth2"
	"github.com/quells/mastobot/internal/toot"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	appName   string
	userEmail string
	password  string

	tootVisibilityS string
	tootVisibility  toot.Visibility
	tootSensitive   bool
	tootSpoilerText string

	maxAge time.Duration
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

	appTootCmd.Flags().StringVar(&tootVisibilityS, "visibility", "private", "[private, unlisted, public, direct]")
	appTootCmd.Flags().BoolVar(&tootSensitive, "sensitive", false, "Mark Toot as containing sensitive material")
	appTootCmd.Flags().StringVar(&tootSpoilerText, "spoiler", "", "Spoiler text")
	appCmd.AddCommand(appTootCmd)

	appExpireCmd.Flags().DurationVar(&maxAge, "max-age", 30*24*time.Hour, "Maximum age")
	appCmd.AddCommand(appExpireCmd)

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

var appTokenRevokeCmd = &cobra.Command{} // TODO

var appTootCmd = &cobra.Command{
	Use:   "toot",
	Short: "Toot!",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("must provide toot message")
		}
		tootVisibility = toot.VisibilityFrom(tootVisibilityS)
		if tootVisibility == toot.VisibilityInvalid {
			return fmt.Errorf("invalid visibility value")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := toot.VerifyCredentials(cmd.Context(), instance, appName)
		if err != nil {
			return err
		}

		status := toot.Status{
			Text:       args[0],
			Visibility: tootVisibility,
			Sensitive:  tootSensitive,
			Spoiler:    tootSpoilerText,
		}
		id, err := status.Submit(cmd.Context(), instance, appName)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, id)
		return nil
	},
}

var appExpireCmd = &cobra.Command{
	Use:   "expire",
	Short: "Delete old toots",
	Long:  "Delete all toots older than a certain age.",
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, err := toot.VerifyCredentials(cmd.Context(), instance, appName)
		if err != nil {
			return err
		}

		list := toot.ListStatuses{
			Limit: 40,
		}
		for {
			statuses, err := list.ForAccount(cmd.Context(), instance, appName, accountID)
			if err != nil {
				return err
			}
			if len(statuses) == 0 {
				break
			}

			for _, status := range statuses {
				if time.Since(status.CreatedAt) < maxAge {
					continue
				}

				err = toot.Delete(cmd.Context(), instance, appName, status.ID)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintln(os.Stdout, status.ID)
			}

			list.MaxID = statuses[len(statuses)-1].ID
		}
		return nil
	},
}
