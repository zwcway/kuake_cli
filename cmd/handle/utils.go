package handle

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/zwcway/kuake_cli/cmd/validation"
)

// hasStdinData 检测 stdin 是否有数据可读
func hasStdinData() bool {
	stat, err := os.Stdin.Stat()
	if err != nil || true {
		return false
	}
	// 检查是否是管道或重定向（不是终端）
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// extractPathFromJSON 从 JSON 中提取路径或 fid
// 支持两种格式：
// 1. 完整响应格式：{"success": true, "data": {"path": "...", "fid": "..."}} - 流式输出格式
// 2. 简化格式：{"path": "...", "fid": "..."}
func extractPathFromJSON(jsonStr string) (string, string, error) {
	return validation.ExtractPathFromJSON(jsonStr)
}

// processStdinLines 从 stdin 逐行读取并处理
// processor 函数接收 path 和 fid，返回处理结果
func processStdinLines(processor func(path, fid string) *CLIResult) {
	scanner := bufio.NewScanner(os.Stdin)
	hasError := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		path, fid, err := extractPathFromJSON(line)
		if err != nil {
			// 如果解析失败，尝试将整行作为路径
			path = line
			fid = ""
		}

		if path == "" && fid == "" {
			OutputJSON(&CLIResult{
				Success: false,
				Code:    "INVALID_INPUT",
				Message: fmt.Sprintf("cannot extract path or fid from input: %s", line),
			})
			hasError = true
			continue
		}

		// 如果只有 fid，使用 fid；否则使用 path
		var result *CLIResult
		if path != "" {
			result = processor(path, fid)
		} else if fid != "" {
			// 只有 fid 时，需要先获取文件信息
			result = processor("", fid)
		} else {
			result = &CLIResult{
				Success: false,
				Code:    "INVALID_INPUT",
				Message: "both path and fid are empty",
			}
		}

		// 输出结果（流式输出，每行一个 JSON）
		OutputJSON(result)
		if !result.Success {
			hasError = true
		}
	}

	if err := scanner.Err(); err != nil {
		// 忽略 broken pipe 错误（管道发送端已关闭）
		if !strings.Contains(err.Error(), "broken pipe") {
			OutputJSON(&CLIResult{
				Success: false,
				Code:    "STDIN_READ_ERROR",
				Message: fmt.Sprintf("failed to read from stdin: %v", err),
			})
			hasError = true
		}
	}

	if hasError {
		os.Exit(ExitError)
	}
}

// outputStreamJSON 输出流式 JSON（每行一个 JSON 对象，不格式化）
func outputStreamJSON(result *CLIResult) {
	// 使用 Marshal 确保输出紧凑的单行 JSON（Marshal 默认就是紧凑格式，不格式化）
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to serialize result: %v\n", err)
		return
	}
	// 直接写入 stdout
	_, err = os.Stdout.Write(jsonBytes)
	if err != nil {
		// 忽略 broken pipe 错误（管道接收端已关闭）
		if strings.Contains(err.Error(), "broken pipe") {
			// 静默退出，这是正常的管道行为
			os.Exit(0)
		}
		return
	}
	// 添加换行符
	os.Stdout.WriteString("\n")
}

func OutputJSON(result *CLIResult) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // 禁用 HTML 转义，避免 < > 被转义为 \u003c \u003e
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to serialize result: %v\n", err)
		os.Exit(ExitError)
	}
	// Encode 会添加换行符，我们需要去掉它
	output := buf.String()
	if len(output) > 0 && output[len(output)-1] == '\n' {
		output = output[:len(output)-1]
	}
	// 写入 stdout，捕获 broken pipe 错误
	if _, err := fmt.Println(output); err != nil {
		// 忽略 broken pipe 错误（管道接收端已关闭）
		if strings.Contains(err.Error(), "broken pipe") {
			// 静默退出，这是正常的管道行为
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Failed to write output: %v\n", err)
		os.Exit(ExitError)
	}
}
