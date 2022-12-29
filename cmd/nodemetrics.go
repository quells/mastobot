package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/quells/mastobot/internal/nodeexporter"
	"github.com/quells/mastobot/internal/toot"
	"github.com/spf13/cobra"
)

var (
	metricsURL string
)

func init() {
	nodemetricsCmd.PersistentFlags().StringVar(&instance, "instance", "", "Mastodon (or compatible) instance to interact with")
	must(nodemetricsCmd.MarkPersistentFlagRequired("instance"))

	nodemetricsCmd.Flags().StringVar(&metricsURL, "metrics-url", "", "URL of the node_exporter metrics")
	must(nodemetricsCmd.MarkFlagRequired("metrics-url"))
	rootCmd.AddCommand(nodemetricsCmd)
}

func nodemetricsToot(m nodeexporter.Metrics) string {
	siBytes := func(f float64) string {
		units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
		si := 0
		for {
			if f < 1024 {
				break
			}
			si++
			f /= 1024
		}
		return fmt.Sprintf("%.2f %s", f, units[si])
	}

	b := new(strings.Builder)
	_, _ = fmt.Fprint(b, m.Name, "\n")
	_, _ = fmt.Fprintf(b, "Load: %.2f %.2f %.2f\n", m.Load1, m.Load5, m.Load15)
	_, _ = fmt.Fprintf(b, "RAM: %s of %s free\n", siBytes(m.MemFree), siBytes(m.MemTotal))
	_, _ = fmt.Fprintf(b, "SWAP: %s of %s free\n", siBytes(m.SwapFree), siBytes(m.SwapTotal))
	_, _ = fmt.Fprintf(b, "Root FS: %s of %s available\n", siBytes(m.RootFSAvail), siBytes(m.RootFSSize))
	_, _ = fmt.Fprintf(b, "Network I/O: %sps | %sps\n", siBytes(m.NetworkIn), siBytes(m.NetworkOut))
	_, _ = fmt.Fprintf(b, "Uptime: %s", m.Uptime)
	return b.String()
}

var nodemetricsCmd = &cobra.Command{
	Use:   "nodemetrics",
	Short: "Node Metrics",
	Long:  `Toots current metrics about the host`,
	RunE: func(cmd *cobra.Command, args []string) error {
		const appName = "nodemetrics"

		_, err := toot.VerifyCredentials(cmd.Context(), instance, appName)
		if err != nil {
			return err
		}

		interval := 5 * time.Second
		metrics, err := nodeexporter.GetMetrics(cmd.Context(), metricsURL, interval)
		if err != nil {
			return err
		}
		status := toot.Status{
			Text:       nodemetricsToot(metrics),
			Visibility: toot.VisibilityPrivate,
		}
		var id string
		id, err = status.Submit(cmd.Context(), instance, "nodemetrics")
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, id)

		return nil
	},
}
