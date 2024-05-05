package taskapi

type TaskRequest struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	ID          string            `json:"id"`
	PortMapping map[string]string `json:"portMapping"`
	Resources   Resources         `json:"resources"`
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
