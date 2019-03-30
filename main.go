package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"path/filepath"
)

var (
	configuration *Configuration
)

func initConfiguration() {
	configuration = &Configuration{}

	err := env.Parse(configuration)
	check(err)
}

func init() {
	initConfiguration()
}

func main() {
	var results []*ResourcesUsage
	container, resourcesStream := listenResourcesUsage()

	for item := range resourcesStream {
		results = append(results, item)
	}
	if len(results) == 0 {
		log.Fatal(
			fmt.Sprintf(
				"Failed to get metrics from container %s. 0 events received",
				configuration.ContainerID,
			),
		)
	}

	saveMetrics(container, results)
}

func listenResourcesUsage() (types.ContainerJSON, <-chan *ResourcesUsage) {
	cli, err := client.NewEnvClient()
	check(err)

	container, err := cli.ContainerInspect(
		context.Background(),
		configuration.ContainerID,
	)
	check(err)

	results, _, err := ListenResourcesUsage(cli, container.ID)
	check(err)

	return container, results
}

func saveMetrics(container types.ContainerJSON, metrics []*ResourcesUsage) {
	directoryPath := filepath.Join(
		configuration.ResultDirectory,
		fmt.Sprintf("%s (%s)", container.Name, container.ID),
	)

	csvReporter := NewCsvReporter(directoryPath)
	pngReporter := NewPngReporter(directoryPath)

	var (
		cpuMetrics    []*Metric
		memoryMetrics []*Metric
	)

	for _, metric := range metrics {
		cpuMetrics = append(cpuMetrics, &Metric{
			Timestamp: metric.CurrentTime.Unix(),
			Value:     metric.CPUUsagePercentage,
		})
		memoryMetrics = append(memoryMetrics, &Metric{
			Timestamp: metric.CurrentTime.Unix(),
			Value:     metric.MemoryUsagePercentage,
		})
	}

	var err error
	err = csvReporter.Report("cpu_percentage", "CPU usage (percentage)", cpuMetrics)
	err = csvReporter.Report("ram_percentage", "RAM usage (percentage)", memoryMetrics)

	err = pngReporter.Report("cpu_percentage", "CPU usage (percentage)", cpuMetrics)
	err = pngReporter.Report("ram_percentage", "RAM usage (percentage)", memoryMetrics)

	check(err)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
