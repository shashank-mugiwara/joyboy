package task

type Local struct {
	Name        string  `json:"name" bson:"name"`
	PortMapping string  `json:"portMapping" bson:"portMapping"`
	Cpus        float32 `json:"cpus" bson:"cpus"`
	Memory      int64   `json:"memory" bson:"memory"`
	ContainerID string  `json:"containerId" bson:"containerId"`
	Owner       string  `json:"owner" bson:"owner"`
}
