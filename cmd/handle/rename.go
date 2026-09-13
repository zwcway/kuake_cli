package handle

import "github.com/zwcway/kuake_cli/sdk"

// Rename 处理重命名命令
func Rename(client *sdk.QuarkClient, args []string) *CLIResult {
	if len(args) < 2 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: `Usage: rename <path> <newName> (all parameters must be quoted, e.g., rename 'file(1).txt' 'new_name.txt')`,
		}
	}

	path := args[0]
	newName := args[1]

	response, err := client.Rename(path, newName)
	if err != nil {
		return &CLIResult{
			Success: false,
			Message: err.Error(),
		}
	}

	if !response.Success {
		return &CLIResult{
			Success: false,
			Code:    response.Code,
			Message: response.Message,
		}
	}

	return &CLIResult{
		Success: true,
		Code:    response.Code,
		Message: response.Message,
		Data:    response.Data,
	}
}
