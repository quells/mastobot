package cmd

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/quells/mastobot/internal/app"
	"github.com/quells/mastobot/internal/nodeexporter"
	"github.com/quells/mastobot/internal/toot"
	"github.com/rs/zerolog/log"
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

func nodemetricsToot(m *nodeexporter.NodeMetrics, prev nodemetricsState) string {
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
	if prev.systemTime != 0 {
		dT := m.TimeSeconds - prev.systemTime
		dRx := m.NetworkReceiveBytes - prev.networkRx
		dTx := m.NetworkTransmitBytes - prev.networkTx
		if dT > 0 && dRx >= 0 && dTx >= 0 {
			dRxdT := uint64(math.Round(float64(dRx) / float64(dT)))
			dTxdT := uint64(math.Round(float64(dTx) / float64(dT)))
			_, _ = fmt.Fprintf(b, "Network I/O: %sps | %sps\n", siBytes(dRxdT), siBytes(dTxdT))
		}
	}
	_, _ = fmt.Fprintf(b, "Uptime: %s", m.Uptime())
	return b.String()
}

type nodemetricsState struct {
	systemTime uint64
	networkRx  uint64
	networkTx  uint64
}

func nodemetricsGetPrevState(ctx context.Context, appName string) (prev nodemetricsState, err error) {
	var systemTime, networkRx, networkTx string
	systemTime, err = app.GetValue(ctx, instance, appName, "systemTime")
	if err != nil {
		return
	}
	if systemTime == "" {
		return
	}
	networkRx, err = app.GetValue(ctx, instance, appName, "networkRx")
	if err != nil {
		return
	}
	networkTx, err = app.GetValue(ctx, instance, appName, "networkTx")
	if err != nil {
		return
	}

	prev.systemTime, _ = strconv.ParseUint(systemTime, 10, 64)
	prev.networkRx, _ = strconv.ParseUint(networkRx, 10, 64)
	prev.networkTx, _ = strconv.ParseUint(networkTx, 10, 64)
	return
}

func nodemetricsSaveState(ctx context.Context, appName string, m *nodeexporter.NodeMetrics) {
	var systemTime, networkRx, networkTx string
	systemTime = strconv.FormatUint(m.TimeSeconds, 10)
	networkRx = strconv.FormatUint(m.NetworkReceiveBytes, 10)
	networkTx = strconv.FormatUint(m.NetworkTransmitBytes, 10)
	if err := app.SetValue(ctx, instance, appName, "systemTime", systemTime); err != nil {
		log.Error().Err(err).Msg("failed to save systemTime state")
	}
	if err := app.SetValue(ctx, instance, appName, "networkRx", networkRx); err != nil {
		log.Error().Err(err).Msg("failed to save networkRx state")
	}
	if err := app.SetValue(ctx, instance, appName, "networkTx", networkTx); err != nil {
		log.Error().Err(err).Msg("failed to save networkTx state")
	}
}

var nodemetricsCmd = &cobra.Command{
	Use:   "nodemetrics",
	Short: "Node Metrics",
	Long:  `Toots current metrics about the host`,
	RunE: func(cmd *cobra.Command, args []string) error {
		const appName = "nodemetrics"
		ctx := cmd.Context()

		metrics, err := nodeexporter.GetNodeMetrics(ctx, metricsURL)
		if err != nil {
			return err
		}

		var prevState nodemetricsState
		prevState, err = nodemetricsGetPrevState(ctx, appName)
		if err != nil {
			return err
		}

		status := toot.Status{
			Text:       nodemetricsToot(metrics, prevState),
			Visibility: toot.VisibilityPrivate,
		}

		if dryRun {
			fmt.Println(status.Text)
			return nil
		}

		_, err = toot.VerifyCredentials(ctx, instance, appName)
		if err != nil {
			return err
		}

		var id string
		id, err = status.Submit(ctx, instance, "nodemetrics")
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(os.Stdout, id)

		nodemetricsSaveState(ctx, appName, metrics)

		return nil
	},
}
