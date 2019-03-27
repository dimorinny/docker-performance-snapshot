package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{},
	)
	if err != nil {
		panic(err)
	}

	container := containers[0]

	results, _, err := ListenResourcesUsage(cli, container.ID)
	if err != nil {
		panic(err)
	}

	for item := range results {
		fmt.Println(item)
	}
}
