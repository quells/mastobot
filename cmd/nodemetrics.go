package cmd

import (
	"fmt"
	"os"
	"strings"

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

func nodemetricsToot(m *nodeexporter.NodeMetrics) string {
	siBytes := func(f uint64) string {
		units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}
		si := 0
		for {
			if f < 1024 {
				break
			}
			si++
			f /= 1024
		}
		return fmt.Sprintf("%d %s", f, units[si])
	}

	b := new(strings.Builder)
	_, _ = fmt.Fprint(b, m.Hostname, "\n")
	if m.Kernel != "" {
		_, _ = fmt.Fprint(b, "Kernel: ", m.Kernel, "\n")
	}
	_, _ = fmt.Fprintf(b, "Load: %.2f %.2f %.2f\n", m.Load1, m.Load5, m.Load15)
	_, _ = fmt.Fprintf(b, "RAM: %s of %s free\n", siBytes(m.MemoryFreeBytes), siBytes(m.MemoryTotalBytes))
	_, _ = fmt.Fprintf(b, "SWAP: %s of %s free\n", siBytes(m.MemorySwapFreeBytes), siBytes(m.MemorySwapTotalBytes))
	_, _ = fmt.Fprintf(b, "Root FS: %s of %s available\n", siBytes(m.FilesystemAvailBytes), siBytes(m.FilesystemSizeBytes))
	//_, _ = fmt.Fprintf(b, "Network I/O: %sps | %sps\n", siBytes(m.NetworkIn), siBytes(m.NetworkOut))
	_, _ = fmt.Fprintf(b, "Uptime: %s", m.Uptime())
	return b.String()
}

var nodemetricsCmd = &cobra.Command{
	Use:   "nodemetrics",
	Short: "Node Metrics",
	Long:  `Toots current metrics about the host`,
	RunE: func(cmd *cobra.Command, args []string) error {
		const appName = "nodemetrics"

		metrics, err := nodeexporter.GetNodeMetrics(cmd.Context(), metricsURL)
		if err != nil {
			return err
		}
		status := toot.Status{
			Text:       nodemetricsToot(metrics),
			Visibility: toot.VisibilityPrivate,
		}

		if dryRun {
			fmt.Println(status.Text)
			return nil
		}

		_, err = toot.VerifyCredentials(cmd.Context(), instance, appName)
		if err != nil {
			return err
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
