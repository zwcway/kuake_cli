package handle

import "github.com/zwcway/kuake_cli/sdk"

// Copy 处理复制命令
func Copy(client *sdk.QuarkClient, args []string) *CLIResult {
	if len(args) < 2 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: `Usage: copy <src> <dest> (all parameters must be quoted, e.g., copy 'file(1).txt' '/dest/')`,
		}
	}

	srcPath := args[0]
	destPath := args[1]

	response, err := client.Copy(srcPath, destPath)
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
