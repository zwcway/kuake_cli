package handle

import "github.com/zwcway/kuake_cli/sdk"

// List 处理列出目录命令
func List(client *sdk.QuarkClient, args []string) *CLIResult {
	dirPath := "/"
	streamMode := false

	// 解析参数，支持 --stream 选项
	var filteredArgs []string
	for i, arg := range args {
		if arg == "--stream" || arg == "-s" {
			streamMode = true
		} else if i == 0 {
			dirPath = arg
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	response, err := client.List(dirPath)
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

	// 流式模式：每行输出一个文件的 JSON
	if streamMode {
		// 从 response.Data 中提取 list 数组
		// response.Data 的类型是 map[string]interface{}，但 list 字段的实际类型是 []sdk.QuarkFileInfo
		if quarkFileInfos, ok := response.Data["list"].([]sdk.QuarkFileInfo); ok {
			// 将 QuarkFileInfo 转换为 map[string]interface{} 并逐行输出
			for _, qfi := range quarkFileInfos {
				fileInfo := map[string]interface{}{
					"fid":          qfi.Fid,
					"file_name":    qfi.Name,
					"path":         qfi.Path,
					"size":         qfi.Size,
					"ctime":        qfi.CreateTime,
					"mtime":        qfi.ModifyTime,
					"dir":          qfi.IsDirectory,
					"download_url": qfi.DownloadURL,
					"created_at":   qfi.CreatedAt,
					"updated_at":   qfi.UpdatedAt,
					"l_created_at": qfi.LCreatedAt,
					"l_updated_at": qfi.LUpdatedAt,
				}
				fileResult := &CLIResult{
					Success: true,
					Code:    response.Code,
					Message: "OK",
					Data:    fileInfo,
				}
				outputStreamJSON(fileResult)
			}
			// 流式模式下不返回结果，已经逐行输出
			return nil
		}
		// 如果无法提取 list，回退到普通模式
	}

	return &CLIResult{
		Success: true,
		Code:    response.Code,
		Message: response.Message,
		Data:    response.Data,
	}
}
