package handle

import (
	"time"

	"github.com/zwcway/kuake_cli/cmd/validation"
	"github.com/zwcway/kuake_cli/sdk"
)

// ShareCreate 处理创建分享链接命令
func ShareCreate(client *sdk.QuarkClient, args []string) *CLIResult {
	if len(args) < 3 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: "Usage: share <path> <days> <passcode> (path and passcode must be quoted, e.g., share \"file(1).txt\" 7 \"false\")",
		}
	}

	path := args[0]

	// 解析有效期天数（必传）
	expireDays, err := validation.ParseIntArg(args[1], "days")
	if err != nil {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: err.Error(),
		}
	}

	// 解析是否需要提取码（必传）
	needPasscode, err := validation.ParseBoolArg(args[2], "passcode")
	if err != nil {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: err.Error(),
		}
	}

	shareInfo, err := client.CreateShare(path, expireDays, needPasscode)
	if err != nil {
		return &CLIResult{
			Success: false,
			Message: err.Error(),
		}
	}

	data := map[string]interface{}{
		"share_url":  shareInfo.ShareURL,
		"pwd_id":     shareInfo.PwdID,
		"passcode":   shareInfo.Passcode,
		"expires_at": shareInfo.ExpiresAt,
	}

	if shareInfo.ExpiresAt > 0 {
		expireTime := time.Unix(shareInfo.ExpiresAt/1000, 0)
		data["expires_at_formatted"] = expireTime.Format("2006-01-02 15:04:05")
	}

	return &CLIResult{
		Success: true,
		Code:    "OK",
		Message: "Share link created successfully",
		Data:    data,
	}
}
