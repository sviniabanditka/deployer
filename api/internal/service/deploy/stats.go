package deploy

import (
	"encoding/json"
	"io"
)

// ContainerStats holds parsed container resource usage.
type ContainerStats struct {
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryUsage uint64  `json:"memory_usage"`
	MemoryLimit uint64  `json:"memory_limit"`
	NetworkRx   uint64  `json:"network_rx"`
	NetworkTx   uint64  `json:"network_tx"`
}

// dockerStats mirrors the JSON structure returned by the Docker stats API.
type dockerStats struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
		OnlineCPUs     uint32 `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemCPUUsage uint64 `json:"system_cpu_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
}

// parseStats decodes a single Docker stats JSON blob and returns a ContainerStats.
func parseStats(r io.Reader) (*ContainerStats, error) {
	var raw dockerStats
	if err := json.NewDecoder(r).Decode(&raw); err != nil {
		return nil, err
	}

	// CPU percentage calculation (same formula as docker CLI).
	cpuDelta := float64(raw.CPUStats.CPUUsage.TotalUsage - raw.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(raw.CPUStats.SystemCPUUsage - raw.PreCPUStats.SystemCPUUsage)

	var cpuPercent float64
	if systemDelta > 0 && cpuDelta > 0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(raw.CPUStats.OnlineCPUs) * 100.0
	}

	var rxBytes, txBytes uint64
	for _, v := range raw.Networks {
		rxBytes += v.RxBytes
		txBytes += v.TxBytes
	}

	return &ContainerStats{
		CPUPercent:  cpuPercent,
		MemoryUsage: raw.MemoryStats.Usage,
		MemoryLimit: raw.MemoryStats.Limit,
		NetworkRx:   rxBytes,
		NetworkTx:   txBytes,
	}, nil
}
