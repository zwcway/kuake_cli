package handle

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/zwcway/kuake_cli/cmd/validation"
	"github.com/zwcway/kuake_cli/sdk"
)

// Upload 处理上传文件命令
func Upload(client *sdk.QuarkClient, args []string) *CLIResult {
	if len(args) < 2 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: `Usage: upload <file> <dest> [--max_upload_parallel N] [--policy skip|overwrite|rsync] (all parameters must be quoted)`,
		}
	}

	filePath := args[0]
	destPath := args[1]
	var uploadParallel string
	opts := &sdk.UploadOptions{
		Policy: sdk.UploadPolicySkip, // 默认跳过
	}

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--max_upload_parallel", "--max-upload-parallel", "--upload-parallel":
			if i+1 >= len(args) {
				return &CLIResult{
					Success: false,
					Code:    "INVALID_ARGS",
					Message: "missing value for --max_upload_parallel",
				}
			}
			value := strings.TrimSpace(args[i+1])
			parallel, err := validation.ParseOptionalIntArg(value, "max_upload_parallel", 1)
			if err != nil || parallel < 1 {
				return &CLIResult{
					Success: false,
					Code:    "INVALID_ARGS",
					Message: "invalid --max_upload_parallel, must be integer >= 1",
				}
			}
			uploadParallel = strconv.Itoa(parallel)
			i++
		case "--policy":
			if i+1 >= len(args) {
				return &CLIResult{
					Success: false,
					Code:    "INVALID_ARGS",
					Message: "missing value for --policy (skip/overwrite/rsync)",
				}
			}
			policyArg := strings.ToLower(strings.TrimSpace(args[i+1]))
			if policyArg != "skip" && policyArg != "overwrite" && policyArg != "rsync" {
				return &CLIResult{
					Success: false,
					Code:    "INVALID_ARGS",
					Message: "invalid --policy value, must be 'skip', 'overwrite', or 'rsync'",
				}
			}
			opts.Policy = sdk.UploadPolicy(policyArg)
			i++
		default:
			return &CLIResult{
				Success: false,
				Code:    "INVALID_ARGS",
				Message: fmt.Sprintf("unknown upload option: %s", args[i]),
			}
		}
	}

	if v := resolveUploadParallelForProcess(uploadParallel); v != "" {
		_ = os.Setenv("KUAKE_UPLOAD_PARALLEL", v)
	}

	// 进度回调，显示上传进度、速度和剩余时间
	progressCallback := func(progress *sdk.UploadProgress) {
		if progress == nil {
			return
		}
		// 输出到 stderr，避免干扰 JSON 输出
		if progress.SpeedStr == "秒传（文件已存在）" {
			// 秒传情况，显示特殊提示
			fmt.Fprintf(os.Stderr, "\r上传进度: %d%% | %s", progress.Progress, progress.SpeedStr)
		} else {
			fmt.Fprintf(os.Stderr, "\r上传进度: %d%% | 速度: %s | 剩余: %s",
				progress.Progress, progress.SpeedStr, progress.RemainingStr)
		}
		if progress.Progress == 100 {
			fmt.Fprintf(os.Stderr, "\n")
		}
	}

	response, err := client.UploadFile(filePath, destPath, progressCallback, opts)
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
