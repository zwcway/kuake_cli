package handle

import "github.com/zwcway/kuake_cli/sdk"

// Delete 处理删除命令
func Delete(client *sdk.QuarkClient, args []string) *CLIResult {
	// 检查是否有 stdin 输入（管道模式）
	if hasStdinData() {
		processStdinLines(func(path, fid string) *CLIResult {
			// 优先使用 path，如果没有则尝试使用 fid 作为路径
			targetPath := path
			if targetPath == "" && fid != "" {
				// 尝试直接使用 fid 作为路径（某些情况下可能有效）
				targetPath = fid
			}

			if targetPath == "" {
				return &CLIResult{
					Success: false,
					Code:    "INVALID_INPUT",
					Message: "cannot determine path or fid from input",
				}
			}

			response, err := client.Delete(targetPath)
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
		})
		// processStdinLines 已经处理了所有输出，返回 nil 表示已完成
		return nil
	}

	// 普通模式：从命令行参数读取
	if len(args) < 1 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: `Usage: delete <path> (path must be quoted, e.g., delete 'file(1).txt') or use pipe mode`,
		}
	}

	path := args[0]
	response, err := client.Delete(path)
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
