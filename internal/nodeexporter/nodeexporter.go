package nodeexporter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

func GetNodeMetrics(ctx context.Context, u string) (*NodeMetrics, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	fmt := expfmt.NewFormat(expfmt.TypeTextPlain)
	req.Header.Set("Accept", string(fmt))

	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var parsed *NodeMetrics
	parsed, err = ParseNodeMetrics(resp.Body, fmt)
	if err != nil {
		return nil, err
	}

	return parsed, nil
}

type NodeMetrics struct {
	Load1                float64
	Load5                float64
	Load15               float64
	FilesystemAvailBytes uint64
	FilesystemSizeBytes  uint64
	MemoryFreeBytes      uint64
	MemoryTotalBytes     uint64
	MemorySwapFreeBytes  uint64
	MemorySwapTotalBytes uint64
	BootTimeSeconds      uint64
	TimeSeconds          uint64
	NetworkReceiveBytes  uint64
	NetworkTransmitBytes uint64
	Hostname             string
	Kernel               string
}

func (m *NodeMetrics) Uptime() string {
	if m == nil {
		return "unknown"
	}
	uptime := time.Duration(m.TimeSeconds-m.BootTimeSeconds) * time.Second
	if uptime < 48*time.Hour {
		return uptime.String()
	}
	upDays := float64(uptime) / float64(24*time.Hour)
	return fmt.Sprintf("%.1fd", upDays)
}

func ParseNodeMetrics(r io.Reader, fmt expfmt.Format) (nodemetrics *NodeMetrics, err error) {
	dec := expfmt.NewDecoder(r, fmt)
	nodemetrics = &NodeMetrics{}
	for {
		var family dto.MetricFamily
		err = dec.Decode(&family)
		if err != nil {
			if err == io.EOF {
				err = nil
				break
			}
			return nil, err
		}

		switch family.GetName() {
		case "node_load1":
			nodemetrics.Load1 = getGauge(&family)
		case "node_load5":
			nodemetrics.Load5 = getGauge(&family)
		case "node_load15":
			nodemetrics.Load15 = getGauge(&family)
		case "node_filesystem_avail_bytes":
			nodemetrics.FilesystemAvailBytes = getFilesystem(&family, "/")
		case "node_filesystem_size_bytes":
			nodemetrics.FilesystemSizeBytes = getFilesystem(&family, "/")
		case "node_memory_MemFree_bytes":
			nodemetrics.MemoryFreeBytes = getGaugeUint64(&family)
		case "node_memory_MemTotal_bytes":
			nodemetrics.MemoryTotalBytes = getGaugeUint64(&family)
		case "node_memory_SwapFree_bytes":
			nodemetrics.MemorySwapFreeBytes = getGaugeUint64(&family)
		case "node_memory_SwapTotal_bytes":
			nodemetrics.MemorySwapTotalBytes = getGaugeUint64(&family)
		case "node_boot_time_seconds":
			nodemetrics.BootTimeSeconds = getGaugeUint64(&family)
		case "node_time_seconds":
			nodemetrics.TimeSeconds = getGaugeUint64(&family)
		case "node_network_receive_bytes_total":
			nodemetrics.NetworkReceiveBytes = getNetwork(&family)
		case "node_network_transmit_bytes_total":
			nodemetrics.NetworkTransmitBytes = getNetwork(&family)
		case "node_uname_info":
			for _, m := range family.GetMetric() {
				for _, lp := range m.GetLabel() {
					switch lp.GetName() {
					case "nodename":
						nodemetrics.Hostname = lp.GetValue()
					case "release":
						nodemetrics.Kernel = lp.GetValue()
					}
				}
			}
		}
	}
	return nodemetrics, nil
}

func getGauge(family *dto.MetricFamily) float64 {
	for _, m := range family.GetMetric() {
		g := m.GetGauge()
		return g.GetValue()
	}
	return 0
}

func getGaugeUint64(family *dto.MetricFamily) uint64 {
	for _, m := range family.GetMetric() {
		g := m.GetGauge()
		return uint64(g.GetValue())
	}
	return 0
}

func getFilesystem(family *dto.MetricFamily, mountpoint string) uint64 {
search:
	for _, m := range family.GetMetric() {
		for _, lp := range m.GetLabel() {
			if lp.GetName() == "mountpoint" {
				if lp.GetValue() != mountpoint {
					continue search
				}
			}
		}
		g := m.GetGauge()
		return uint64(g.GetValue())
	}
	return 0
}

func getNetwork(family *dto.MetricFamily) uint64 {
search:
	for _, m := range family.GetMetric() {
		for _, lp := range m.GetLabel() {
			if lp.GetName() != "device" {
				continue
			}
			device := lp.GetValue()
			if strings.HasPrefix(device, "br-") ||
				strings.HasPrefix(device, "docker") ||
				strings.HasPrefix(device, "lo") ||
				strings.HasPrefix(device, "veth") {
				continue search
			}
		}
		c := m.GetCounter()
		return uint64(c.GetValue())
	}
	return 0
}
