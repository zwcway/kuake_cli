package handle

import (
	"fmt"
	"strings"

	"github.com/zwcway/kuake_cli/sdk"
)

// ShareDelete 处理取消分享命令
// 支持两种方式：
// 1. 直接提供 share_id: share-delete "fdd8bfd93f21491ab80122538bec310d"
// 2. 提供文件路径: share-delete "/file.txt" (会先获取文件信息，然后从分享列表中查找share_id)
func ShareDelete(client *sdk.QuarkClient, args []string) *CLIResult {
	if len(args) < 1 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: `Usage: share-delete <share_id_or_path> [share_id_or_path2] ... (e.g., share-delete "fdd8bfd93f21491ab80122538bec310d" or share-delete "/file.txt")`,
		}
	}

	var shareIDs []string
	var paths []string

	// 区分 share_id 和文件路径
	// share_id 通常是32位十六进制字符串，不以 "/" 开头
	// 文件路径通常以 "/" 开头
	for _, arg := range args {
		if strings.HasPrefix(arg, "/") {
			// 是文件路径
			paths = append(paths, arg)
		} else {
			// 假设是 share_id
			shareIDs = append(shareIDs, arg)
		}
	}

	// 处理文件路径：获取文件信息，然后从分享列表中查找share_id
	if len(paths) > 0 {
		for _, path := range paths {
			// 获取文件信息
			fileInfo, err := client.GetFileInfo(path)
			if err != nil {
				return &CLIResult{
					Success: false,
					Code:    "GET_FILE_INFO_ERROR",
					Message: fmt.Sprintf("failed to get file info for path '%s': %v", path, err),
				}
			}

			if !fileInfo.Success {
				return &CLIResult{
					Success: false,
					Code:    fileInfo.Code,
					Message: fmt.Sprintf("failed to get file info for path '%s': %s", path, fileInfo.Message),
				}
			}

			// 获取fid
			fid, ok := fileInfo.Data["fid"].(string)
			if !ok || fid == "" {
				return &CLIResult{
					Success: false,
					Code:    "INVALID_FILE_INFO",
					Message: fmt.Sprintf("file '%s' does not have valid fid", path),
				}
			}

			// 从分享列表中查找share_id
			shareID, err := client.GetShareIDByFid(fid)
			if err != nil {
				return &CLIResult{
					Success: false,
					Code:    "GET_SHARE_ID_ERROR",
					Message: fmt.Sprintf("failed to get share_id for file '%s' (fid: %s): %v. The file may not be shared.", path, fid, err),
				}
			}

			shareIDs = append(shareIDs, shareID)
		}
	}

	// 如果没有找到任何 share_id，返回错误
	if len(shareIDs) == 0 {
		return &CLIResult{
			Success: false,
			Code:    "NO_SHARE_IDS",
			Message: "no valid share_ids found. Please provide share_id(s) or file path(s) with active shares.",
		}
	}

	// 删除分享
	err := client.DeleteShare(shareIDs)
	if err != nil {
		return &CLIResult{
			Success: false,
			Message: err.Error(),
		}
	}

	resultData := map[string]interface{}{
		"deleted_share_ids": shareIDs,
	}
	if len(paths) > 0 {
		resultData["processed_paths"] = paths
	}

	return &CLIResult{
		Success: true,
		Code:    "OK",
		Message: "Share deleted successfully",
		Data:    resultData,
	}
}
