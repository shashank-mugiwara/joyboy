package taskapi

type TaskRequest struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	ID    string `json:"id"`
}

type TaskResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
	State string `json:"state"`
}