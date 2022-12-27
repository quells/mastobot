package nodeexporter

import (
	"context"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type Metrics struct {
	Name                    string
	Load1, Load5, Load15    float64
	RootFSAvail, RootFSSize float64
	MemFree, MemTotal       float64
	SwapFree, SwapTotal     float64
	Uptime                  time.Duration
	NetworkIn, NetworkOut   float64
}

func GetMetrics(ctx context.Context, u string, interval time.Duration) (metrics Metrics, err error) {
	var a, b map[string]any
	a, err = getMetrics(ctx, u)
	if err != nil {
		return
	}

	t := time.NewTimer(interval)
	select {
	case <-ctx.Done():
		if !t.Stop() {
			<-t.C
		}
		err = ctx.Err()
		return

	case <-t.C:
		break
	}

	b, err = getMetrics(ctx, u)
	if err != nil {
		return
	}

	metrics.Name = b["node_uname_info"].(string)

	metrics.Load1 = (a["load1"].(float64) + b["load1"].(float64)) / 2
	metrics.Load5 = (a["load5"].(float64) + b["load5"].(float64)) / 2
	metrics.Load15 = (a["load15"].(float64) + b["load15"].(float64)) / 2

	metrics.RootFSAvail = (a["node_filesystem_avail_bytes"].(float64) + b["node_filesystem_avail_bytes"].(float64)) / 2
	metrics.RootFSSize = (a["node_filesystem_size_bytes"].(float64) + b["node_filesystem_size_bytes"].(float64)) / 2

	metrics.MemFree = (a["node_memory_MemFree_bytes"].(float64) + b["node_memory_MemFree_bytes"].(float64)) / 2
	metrics.MemTotal = (a["node_memory_MemTotal_bytes"].(float64) + b["node_memory_MemTotal_bytes"].(float64)) / 2
	metrics.SwapFree = (a["node_memory_SwapFree_bytes"].(float64) + b["node_memory_SwapFree_bytes"].(float64)) / 2
	metrics.SwapTotal = (a["node_memory_SwapTotal_bytes"].(float64) + b["node_memory_SwapTotal_bytes"].(float64)) / 2

	uptime := b["node_time_seconds"].(float64) - b["node_boot_time_seconds"].(float64)
	metrics.Uptime = time.Duration(uptime) * time.Second

	metrics.NetworkIn = (b["node_network_receive_bytes_total"].(float64) - a["node_network_receive_bytes_total"].(float64)) / interval.Seconds()
	metrics.NetworkOut = (b["node_network_transmit_bytes_total"].(float64) - a["node_network_transmit_bytes_total"].(float64)) / interval.Seconds()

	return
}

func getMetrics(ctx context.Context, u string) (metrics map[string]any, err error) {
	format := expfmt.FmtProtoDelim
	var req *http.Request
	req, err = http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", string(format))
	req = req.WithContext(ctx)

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	metrics = make(map[string]any)
	for _, name := range []string{
		"load1",
		"load5",
		"load15",
		"node_filesystem_avail_bytes",
		"node_filesystem_size_bytes",
		"node_memory_MemFree_bytes",
		"node_memory_MemTotal_bytes",
		"node_memory_SwapFree_bytes",
		"node_memory_SwapTotal_bytes",
		"node_boot_time_seconds",
		"node_time_seconds",
		"node_network_receive_bytes_total",
		"node_network_transmit_bytes_total",
		"node_uname_info",
	} {
		metrics[name] = float64(0)
	}

	for {
		var family dto.MetricFamily
		err = expfmt.NewDecoder(resp.Body, format).Decode(&family)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			log.Debug().Msgf("%+v", metrics)
			return
		}
		name := *family.Name
		if _, ok := metrics[name]; !ok {
			continue
		}

		for _, m := range family.Metric {
			labels := make(map[string]string)
			for _, label := range m.GetLabel() {
				labels[label.GetName()] = label.GetValue()
			}
			if strings.Contains(name, "network") {
				device := labels["device"]
				if strings.HasPrefix(device, "br-") ||
					strings.HasPrefix(device, "docker") ||
					strings.HasPrefix(device, "lo") ||
					strings.HasPrefix(device, "veth") {
					continue
				}
			} else if strings.Contains(name, "filesystem") {
				mountpoint := labels["mountpoint"]
				if mountpoint != "/" {
					continue
				}
			}

			if name == "node_uname_info" {
				metrics[name] = labels["nodename"]
				continue
			}

			if g := m.GetGauge(); g != nil {
				metrics[name] = g.GetValue()
			} else if c := m.GetCounter(); c != nil {
				metrics[name] = c.GetValue()
			}
		}
	}
}
