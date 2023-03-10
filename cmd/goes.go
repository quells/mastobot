package cmd

import (
	"fmt"
	"os"

	"github.com/quells/mastobot/internal/goes"
	"github.com/quells/mastobot/internal/toot"
	"github.com/spf13/cobra"
)

func init() {
	goesCmd.PersistentFlags().StringVar(&instance, "instance", "", "Mastodon (or compatible) instance to interact with")
	must(goesCmd.MarkPersistentFlagRequired("instance"))

	goesCmd.AddCommand(goesWestCmd)
	rootCmd.AddCommand(goesCmd)
}

var goesCmd = &cobra.Command{
	Use:   "goes",
	Short: "Toot satellite image of Earth's hemisphere",
}

var goesWestCmd = &cobra.Command{
	Use:   "west",
	Short: "Toot satellite image of Earth's western hemisphere from GOES-17",
	RunE: func(cmd *cobra.Command, args []string) error {
		const appName = "GOES-17"

		_, err := toot.VerifyCredentials(cmd.Context(), instance, appName)
		if err != nil {
			return err
		}

		var large, thumbnail []byte
		large, thumbnail, err = goes.GOES17(cmd.Context())
		if err != nil {
			return err
		}

		upload := toot.MediaUpload{
			ContentType: toot.ContentTypeMediaJPEG,
			File:        large,
			Thumbnail:   thumbnail,
			Description: "Satellite image of the western hemisphere of Earth",
			Focus:       [2]float64{0.5, 0.5},
		}
		var mediaID string
		mediaID, err = upload.Submit(cmd.Context(), instance, appName)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, mediaID)

		status := toot.Status{
			MediaIDs:   []string{mediaID},
			Visibility: toot.VisibilityPublic,
		}
		var statusID string
		statusID, err = status.Submit(cmd.Context(), instance, appName)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, statusID)

		return nil
	},
}
