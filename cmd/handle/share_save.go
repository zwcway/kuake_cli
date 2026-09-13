package handle

import (
	"fmt"
	"strings"

	"github.com/zwcway/kuake_cli/sdk"
)

// ShareSave 处理转存分享文件命令
// 用法: share-save <share_link> [passcode] [dest_dir]
func ShareSave(client *sdk.QuarkClient, args []string) *CLIResult {
	if len(args) < 1 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: `Usage: share-save <share_link> [passcode] [dest_dir] (e.g., share-save "https://pan.quark.cn/s/xxx" "1234" "/folder")`,
		}
	}

	shareLink := args[0]
	var passcode string
	var destDir string

	// 解析参数
	if len(args) >= 2 {
		// 第二个参数可能是 passcode 或 dest_dir（如果以 / 开头）
		if strings.HasPrefix(args[1], "/") {
			destDir = args[1]
		} else {
			passcode = args[1]
		}
	}
	if len(args) >= 3 {
		destDir = args[2]
	}

	// 从分享链接中提取 pwdID 和 passcode
	shareInfo, err := client.GetShareInfo(shareLink)
	if err != nil {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_SHARE_LINK",
			Message: fmt.Sprintf("failed to parse share link: %v", err),
		}
	}

	// 如果命令行提供了 passcode，优先使用命令行的
	if passcode == "" && shareInfo.Passcode != "" {
		passcode = shareInfo.Passcode
	}

	// 获取 stoken
	stokenData, err := client.GetShareStoken(shareInfo.PwdID, passcode)
	if err != nil {
		return &CLIResult{
			Success: false,
			Code:    "GET_STOKEN_ERROR",
			Message: fmt.Sprintf("failed to get share stoken: %v", err),
		}
	}

	// 从 stokenData 中提取 stoken
	stoken, ok := stokenData["stoken"].(string)
	if !ok || stoken == "" {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_STOKEN",
			Message: "stoken not found in response",
		}
	}

	// 处理目标目录
	toPdirFid := "0" // 默认根目录
	if destDir != "" {
		if destDir == "/" {
			toPdirFid = "0"
		} else if strings.HasPrefix(destDir, "/") {
			// 是路径，需要转换为 FID
			dirInfo, err := client.GetFileInfo(destDir)
			if err != nil {
				return &CLIResult{
					Success: false,
					Code:    "GET_DEST_DIR_ERROR",
					Message: fmt.Sprintf("failed to get destination directory info: %v", err),
				}
			}
			if !dirInfo.Success {
				return &CLIResult{
					Success: false,
					Code:    dirInfo.Code,
					Message: fmt.Sprintf("failed to get destination directory: %s", dirInfo.Message),
				}
			}
			// 安全地获取 fid
			fid, ok := dirInfo.Data["fid"].(string)
			if !ok || fid == "" {
				return &CLIResult{
					Success: false,
					Code:    "INVALID_DEST_DIR",
					Message: "destination directory info is invalid: fid not found or empty",
				}
			}
			toPdirFid = fid
		} else {
			// 假设是 FID
			toPdirFid = destDir
		}
	}

	// 转存文件（全部保存）
	// fidList 和 shareTokenList 为空表示全部保存
	result, err := client.SaveShareFile(shareInfo.PwdID, stoken, []string{}, []string{}, toPdirFid, true)
	if err != nil {
		return &CLIResult{
			Success: false,
			Code:    "SAVE_SHARE_ERROR",
			Message: fmt.Sprintf("failed to save share files: %v", err),
		}
	}

	// 构建返回数据
	data := map[string]interface{}{
		"pwd_id":    shareInfo.PwdID,
		"dest_dir":  destDir,
		"dest_fid":  toPdirFid,
		"save_all":  true,
		"save_data": result,
	}

	return &CLIResult{
		Success: true,
		Code:    "OK",
		Message: "Share files saved successfully",
		Data:    data,
	}
}
