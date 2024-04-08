package taskapi

type TaskRequest struct {
	Name  string `json:"taskName"`
	Image string `json:"dockerImage"`
}
