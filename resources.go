package main

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"strings"
)

func ListenResourcesUsage(client *client.Client, containerID string) (<-chan interface{}, <-chan error, error) {
	apiResponse, err := client.ContainerStats(
		context.Background(),
		containerID,
		true,
	)
	if err != nil {
		return nil, nil, err
	}

	resultChannel := make(chan interface{})
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

			resultChannel <- stats
		}

		_ = apiResponse.Body.Close()
	}()

	return resultChannel, errorsChannel, nil
}

// https://github.com/moby/moby/blob/eb131c5383db8cac633919f82abad86c99bffbe5/cli/command/container/stats_helpers.go#L190
func calculateCPUPercentWindows(v *types.StatsJSON) float64 {
	// Max number of 100ns intervals between the previous time read and now
	possIntervals := uint64(v.Read.Sub(v.PreRead).Nanoseconds()) // Start with number of ns intervals
	possIntervals /= 100                                         // Convert to number of 100ns intervals
	possIntervals *= uint64(v.NumProcs)                          // Multiple by the number of processors

	// Intervals used
	intervalsUsed := v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage

	// Percentage avoiding divide-by-zero
	if possIntervals > 0 {
		return float64(intervalsUsed) / float64(possIntervals) * 100.0
	}
	return 0.00
}

// https://github.com/moby/moby/blob/eb131c5383db8cac633919f82abad86c99bffbe5/cli/command/container/stats_helpers.go#L206
func calculateBlockIO(blkio types.BlkioStats) (blkRead uint64, blkWrite uint64) {
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + bioEntry.Value
		case "write":
			blkWrite = blkWrite + bioEntry.Value
		}
	}
	return
}

// https://github.com/moby/moby/blob/eb131c5383db8cac633919f82abad86c99bffbe5/cli/command/container/stats_helpers.go#L218
func calculateNetwork(network map[string]types.NetworkStats) (float64, float64) {
	var rx, tx float64

	for _, v := range network {
		rx += float64(v.RxBytes)
		tx += float64(v.TxBytes)
	}
	return rx, tx
}
