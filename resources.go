package main

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"io"
	"strings"
	"time"
)

type ResourcesUsage struct {
	ContainerID string

	CurrentTime time.Time

	BlockRead,
	BlockWrite float64

	BytesReceived,
	BytesSent float64

	CPUUsagePercentage    float64
	MemoryUsagePercentage float64
}

func ListenResourcesUsage(client *client.Client, containerID string) (<-chan *ResourcesUsage, <-chan error, error) {
	apiResponse, err := client.ContainerStats(
		context.Background(),
		containerID,
		true,
	)
	if err != nil {
		return nil, nil, err
	}
	if apiResponse.OSType == "windows" {
		return nil, nil, errors.New("Unsupported OS: windows")
	}

	resultChannel := make(chan *ResourcesUsage)
	errorsChannel := make(chan error)

	go func() {
		jsonDecoder := json.NewDecoder(apiResponse.Body)

		for {
			var stats *types.StatsJSON

			err := jsonDecoder.Decode(&stats)
			if err != nil {
				// io.EOF in most of cases means container completed (or removed) event
				if err != io.EOF {
					errorsChannel <- err
				}

				close(resultChannel)
				close(errorsChannel)

				break
			}

			cpuPercentageUsage := calculateCPUPercentUnix(stats)
			memoryPercentageUsage := calculateMemoryUsage(stats)
			blockRead, blockWrite := calculateBlockIO(stats.BlkioStats)
			bytesReceived, bytesSent := calculateNetwork(stats.Networks)

			resultChannel <- &ResourcesUsage{
				ContainerID: containerID,

				CurrentTime: time.Now(),

				CPUUsagePercentage:    cpuPercentageUsage,
				MemoryUsagePercentage: memoryPercentageUsage,

				BlockRead:  float64(blockRead),
				BlockWrite: float64(blockWrite),

				BytesReceived: bytesReceived,
				BytesSent:     bytesSent,
			}
		}

		_ = apiResponse.Body.Close()
	}()

	return resultChannel, errorsChannel, nil
}

// based on https://github.com/moby/moby/blob/eb131c5383db8cac633919f82abad86c99bffbe5/cli/command/container/stats_helpers.go#L175
func calculateCPUPercentUnix(stats *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
		// calculate the change for the entire system between readings
		systemDelta = float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(stats.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

// based on https://github.com/moby/moby/blob/eb131c5383db8cac633919f82abad86c99bffbe5/cli/command/container/stats_helpers.go#L107
func calculateMemoryUsage(stats *types.StatsJSON) float64 {
	// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
	// got any data from cgroup
	if stats.MemoryStats.Limit != 0 {
		return float64(stats.MemoryStats.Usage) / float64(stats.MemoryStats.Limit) * 100.0
	}

	return 0
}

// based on https://github.com/moby/moby/blob/eb131c5383db8cac633919f82abad86c99bffbe5/cli/command/container/stats_helpers.go#L206
func calculateBlockIO(blockIO types.BlkioStats) (blockRead uint64, blockWrite uint64) {
	for _, bioEntry := range blockIO.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blockRead = blockRead + bioEntry.Value
		case "write":
			blockWrite = blockWrite + bioEntry.Value
		}
	}
	return
}

// based on https://github.com/moby/moby/blob/eb131c5383db8cac633919f82abad86c99bffbe5/cli/command/container/stats_helpers.go#L218
func calculateNetwork(network map[string]types.NetworkStats) (bytesReceived float64, bytesSent float64) {
	for _, v := range network {
		bytesReceived += float64(v.RxBytes)
		bytesSent += float64(v.TxBytes)
	}
	return
}
