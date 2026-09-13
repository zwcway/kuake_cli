package handle

import "github.com/zwcway/kuake_cli/sdk"

// UserInfo 处理获取用户信息命令
func UserInfo(client *sdk.QuarkClient) *CLIResult {
	response, err := client.GetUserInfo()
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
