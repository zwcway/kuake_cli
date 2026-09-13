package handle

const (
	ExitSuccess = 0
	ExitError   = 1
)

type CLIResult struct {
	Success bool                   `json:"success"`
	Code    string                 `json:"code,omitempty"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}
