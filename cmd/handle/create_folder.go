package handle

import (
	"fmt"
	"strings"

	"github.com/zwcway/kuake_cli/sdk"
)

// CreateFolder 处理创建文件夹命令
func CreateFolder(client *sdk.QuarkClient, args []string) *CLIResult {
	if len(args) < 2 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: `Usage: create <name> <pdir> (all parameters must be quoted, e.g., create 'folder(1)' '/')`,
		}
	}

	folderName := args[0]
	pdirArg := args[1]

	// 处理父目录参数：如果是路径（以 / 开头），需要转换为 FID
	var pdirFid string
	if pdirArg == "" || pdirArg == "/" {
		pdirFid = "/" // 根目录使用标准表示 "/"，SDK 会自动转换为 "0"
	} else if strings.HasPrefix(pdirArg, "/") {
		// 是路径字符串，需要转换为 FID
		dirInfo, err := client.GetFileInfo(pdirArg)
		if err != nil {
			return &CLIResult{
				Success: false,
				Code:    "GET_PARENT_DIRECTORY_ERROR",
				Message: fmt.Sprintf("failed to get parent directory info: %v", err),
			}
		}
		if !dirInfo.Success {
			return &CLIResult{
				Success: false,
				Code:    dirInfo.Code,
				Message: fmt.Sprintf("failed to get parent directory: %s", dirInfo.Message),
			}
		}
		// 安全地获取 fid
		fid, ok := dirInfo.Data["fid"].(string)
		if !ok || fid == "" {
			return &CLIResult{
				Success: false,
				Code:    "INVALID_PARENT_DIRECTORY",
				Message: "parent directory info is invalid: fid not found or empty",
			}
		}
		pdirFid = fid
	} else {
		// 假设是 FID（不是以 / 开头的字符串）
		pdirFid = pdirArg
	}

	response, err := client.CreateFolder(folderName, pdirFid)
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
