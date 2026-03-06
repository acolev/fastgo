package dto

type ProbeResponse struct {
	Status   string         `json:"status"`
	Message  string         `json:"message,omitempty"`
	Services map[string]any `json:"services,omitempty"`
}
