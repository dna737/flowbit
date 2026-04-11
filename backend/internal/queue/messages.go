package queue

type JobMessage struct {
	JobID      string         `json:"job_id"`
	JobType    string         `json:"job_type"`
	Parameters map[string]any `json:"parameters"`
}
