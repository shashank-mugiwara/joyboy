package task

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
	"github.com/shashank-mugiwara/joyboy/config"
	"github.com/shashank-mugiwara/joyboy/database"
	"github.com/shashank-mugiwara/joyboy/utils"
)

type State int

const (
	Pending State = iota
	Scheduled
	Running
	Completed
	Failed
	Stopped
)

func (s State) String() string {
	switch s {
	case Pending:
		return "Pending"
	case Scheduled:
		return "Scheduled"
	case Running:
		return "Running"
	case Completed:
		return "Completed"
	case Failed:
		return "Failed"
	case Stopped:
		return "Stopped"
	default:
		return fmt.Sprintf("Unknown state: %d", s)
	}
}

var KnownContainerStateMap = map[string]string{
	"Pending":   "Pending",
	"Scheduled": "Scheduled",
	"Running":   "Running",
	"Completed": "Completed",
	"Failed":    "Failed",
	"Stopped":   "Stopped",
}

var stateTransitionMap = map[string][]string{
	"Pending":   {"Scheduled"},
	"Scheduled": {"Scheduled", "Running", "Failed"},
	"Running":   {"Running", "Completed", "Failed"},
	"Completed": {},
	"Failed":    {},
}

type Task struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	State         string    `json:"state"`
	Image         string    `json:"image"`
	Memory        int64     `json:"memory"`
	Disk          int64     `json:"disk"`
	ExposedPorts  string    `json:"exposedPorts"`
	PortBindings  string    `json:"portBindings"`
	RestartPolicy string    `json:"restartPolicy"`
	StartTime     time.Time `json:"startTime"`
	EndTime       time.Time `json:"endTime"`
	FinishTime    time.Time `json:"finishTime"`
	Duration      time.Time `json:"duration"`
	ContainerID   string    `json:"containerId"`
	Cpus          float32   `json:"cpus"`
}

type TaskEvent struct {
	ID        uuid.UUID
	State     State
	Timestamp time.Time
	Task      Task
}

type Docker struct {
	Client      *client.Client
	Config      config.Config
	ContainerId string
}

type DockerResult struct {
	Error       error
	Action      string
	ContainerId string
	Result      string
	Message     string
}

func (d *Docker) Run() DockerResult {
	ctx := context.Background()
	reader, err := d.Client.ImagePull(
		ctx, d.Config.Image, image.PullOptions{})

	if err != nil {
		log.Printf("Error pulling the image %s: %v\n", d.Config.Image, d.Config)
		return DockerResult{Error: err}
	}

	defer reader.Close()

	io.Copy(os.Stdout, reader)

	restartPolicy := container.RestartPolicy{
		Name: container.RestartPolicyMode(d.Config.RestartPolicy),
	}

	resources := container.Resources{
		Memory:   d.Config.Memory * 1024 * 1024,
		NanoCPUs: int64(d.Config.Cpus * 1e9),
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(d.Config.PortBindings), &result)
	if err != nil {
		log.Fatalf("Error parsing JSON: %s", err)
		return DockerResult{Error: err}
	}

	json_string, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error parsing JSON: %s", err)
		return DockerResult{Error: err}
	}

	portBindings, err := utils.CreatePortBindings(string(json_string))
	if err != nil {
		log.Printf("Error parsing PortBindings: %v\n", err)
		return DockerResult{Error: err}
	}

	hostConfig := container.HostConfig{
		RestartPolicy:   restartPolicy,
		Resources:       resources,
		PortBindings:    portBindings,
		PublishAllPorts: false,
	}

	exposed_ports, err := utils.CreateExposedPorts(result)
	if err != nil {
		log.Printf("Error parsing PortBindings: %v\n", d.Config.PortBindings)
		return DockerResult{Error: err}
	}

	containerConfig := container.Config{
		Image:        d.Config.Image,
		Env:          d.Config.Env,
		ExposedPorts: exposed_ports,
	}

	resp, err := d.Client.ContainerCreate(
		ctx, &containerConfig, &hostConfig, nil, nil, d.Config.Name)

	if err != nil {
		log.Printf("Error starting container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	err = d.Client.ContainerStart(
		ctx, resp.ID, container.StartOptions{})
	if err != nil {
		log.Printf("Error starting container %s: %v\n", resp.ID, err)
		return DockerResult{Error: err}
	}

	d.ContainerId = resp.ID
	out, err := d.Client.ContainerLogs(
		ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})

	if err != nil {
		log.Printf("Error getting logs for container %s: %v\n", resp.ID, err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return DockerResult{
		ContainerId: resp.ID,
		Action:      "start",
		Result:      "success",
	}
}

func (d *Docker) Stop(id string) DockerResult {
	log.Printf("Attempting to stop container: %v", id)
	ctx := context.Background()
	err := d.Client.ContainerStop(ctx, id, container.StopOptions{})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	err = d.Client.ContainerRemove(ctx, id, container.RemoveOptions{})
	if err != nil {
		panic(err)
	}

	return DockerResult{Action: "stop", Result: "success", Error: nil}
}

func (t *Task) NewConfig(task *Task) config.Config {
	return config.Config{
		Name:         task.Name,
		Image:        task.Image,
		Memory:       int64(task.Memory),
		PortBindings: task.PortBindings,
	}
}

func (t *Task) NewDocker(conf config.Config) Docker {
	dc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		fmt.Printf("error creating docker client %v\n", err)
		panic(err)
	}

	return Docker{
		Config: conf,
		Client: dc,
	}
}

func Contains(states []string, state string) bool {
	for _, s := range states {
		if s == state {
			return true
		}
	}

	return false
}

func ValidStateTransition(src string, dst string) bool {
	return Contains(stateTransitionMap[src], dst)
}

func StopAllTasks() {
	ctx := context.Background()
	dc, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		fmt.Printf("error creating docker client %v\n", err)
		panic(err)
	}

	containers, err := dc.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		panic(err)
	}

	for _, cntr := range containers {
		fmt.Print("Stopping container ", cntr.ID[:10], "... ")
		noWaitTimeout := 0
		if err := dc.ContainerStop(ctx, cntr.ID, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
			panic(err)
		}

		if err := dc.ContainerRemove(ctx, cntr.ID, container.RemoveOptions{RemoveVolumes: true, Force: true}); err != nil {
			panic(err)
		}

		// Update DB entry
		database.GetDb().Model(&Task{}).Where("container_id = ?", cntr.ID).Update("state", Stopped.String())
	}
}
