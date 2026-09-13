package handle

import (
	"github.com/zwcway/kuake_cli/cmd/validation"
	"github.com/zwcway/kuake_cli/sdk"
)

// ShareList 处理获取我的分享列表命令
func ShareList(client *sdk.QuarkClient, args []string) *CLIResult {
	// 解析参数，支持可选参数
	page := 1
	size := 50
	orderField := "created_at"
	orderType := "desc"

	if len(args) > 0 {
		if p, err := validation.ParseOptionalIntArg(args[0], "page", page); err == nil && p > 0 {
			page = p
		}
	}
	if len(args) > 1 {
		if s, err := validation.ParseOptionalIntArg(args[1], "size", size); err == nil && s > 0 {
			size = s
		}
	}
	if len(args) > 2 {
		orderField = args[2]
	}
	if len(args) > 3 {
		orderType = args[3]
	}

	shareList, err := client.GetMyShareList(page, size, orderField, orderType)
	if err != nil {
		return &CLIResult{
			Success: false,
			Message: err.Error(),
		}
	}

	return &CLIResult{
		Success: true,
		Code:    "OK",
		Message: "Get share list successfully",
		Data:    shareList,
	}
}