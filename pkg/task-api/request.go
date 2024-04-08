package taskapi

type TaskRequest struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}
