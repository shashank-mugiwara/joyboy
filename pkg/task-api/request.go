package taskapi

import "github.com/docker/docker/api/types"

type TaskRequest struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	ID          string            `json:"id"`
	PortMapping map[string]string `json:"portMapping"`
	Resources   Resources         `json:"resources"`
	ScaleConfig ScaleConfig       `json:"scaleConfig"`
}

type TaskResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
	State string `json:"state"`
}

type Resources struct {
	Memory int64   `json:"memory"`
	Cpus   float32 `json:"cpus"`
}

type ScaleConfig struct {
	TaskName        string `json:"taskName"`
	MinTaskScale    int    `json:"minTaskScale"`
	MaxTaskScale    int    `json:"maxTaskScale"`
	ScalingStrategy string `json:"scalingStrategy"`
}

type TaskMetrics struct {
	ContainerId string          `json:"containerId"`
	TaskName    string          `json:"taskName"`
	Stats       types.StatsJSON `json:"stats"`
}
