package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/shashank-mugiwara/joyboy/dkrclient"
)

func PollContainerMetrics(containerID string) (types.StatsJSON, error) {
	ctx := context.Background()
	dockerClient := dkrclient.GetPlainDockerClient()

	stats, err := dockerClient.ContainerStats(ctx, containerID, false)
	if err != nil {
		log.Fatalf("Error getting container stats: %s", err)
	}
	defer stats.Body.Close()

	var stat types.StatsJSON
	decoder := json.NewDecoder(stats.Body)
	if err = decoder.Decode(&stat); err != nil {
		if err == io.EOF {
			fmt.Println("No more data in stats stream.")
		} else {
			log.Fatalf("Error decoding stats: %s", err)
		}
	} else {
		fmt.Printf("Container: %s\n", containerID)
		fmt.Printf("CPU Usage: %v\n", stat.CPUStats.CPUUsage.TotalUsage)
		fmt.Printf("Memory Usage: %v bytes\n", stat.MemoryStats.Usage)
		fmt.Printf("Network I/O: %+v\n", stat.Networks)
	}

	return stat, err
}
