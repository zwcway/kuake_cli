package main

import (
	"fmt"
	"os"

	"github.com/zwcway/kuake_cli/cmd/handle"
	"github.com/zwcway/kuake_cli/sdk"
)

// Version 版本号，与编译产物名称一致
var Version = "v1.5.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(handle.ExitError)
	}

	// 解析命令行参数，支持 -cookies 参数
	var cookies string
	var command string
	var args []string
	skipNext := false

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]

		if skipNext {
			skipNext = false
			continue
		}

		// 检查是否是 cookies 参数
		if arg == "-cookies" || arg == "--cookies" {
			if i+1 < len(os.Args) {
				cookies = os.Args[i+1]
				skipNext = true
				continue
			} else {
				handle.OutputJSON(&handle.CLIResult{
					Success: false,
					Code:    "INVALID_ARGS",
					Message: fmt.Sprintf("%s requires a cookies value", arg),
				})
				os.Exit(handle.ExitError)
			}
		}

		// 第一个非配置参数是命令
		if command == "" {
			// 检查是否是帮助命令
			if arg == "help" || arg == "-h" || arg == "--help" {
				printUsage()
				os.Exit(handle.ExitSuccess)
			}
			// 检查是否是版本命令（在 QuarkClient 初始化之前拦截，无需配置文件）
			if arg == "version" || arg == "-v" || arg == "--version" {
				handle.OutputJSON(&handle.CLIResult{
					Success: true,
					Code:    "OK",
					Message: fmt.Sprintf("kuake %s", Version),
					Data: map[string]interface{}{
						"version": Version,
					},
				})
				os.Exit(handle.ExitSuccess)
			}
			command = arg
		} else {
			args = append(args, arg)
		}
	}

	if command == "" {
		printUsage()
		os.Exit(handle.ExitError)
	}

	loadDotEnvFiles()

	// 创建客户端
	var client *sdk.QuarkClient
	defer func() {
		if r := recover(); r != nil {
			handle.OutputJSON(&handle.CLIResult{
				Success: false,
				Code:    "INIT_ERROR",
				Message: fmt.Sprintf("Failed to initialize client: %v", r),
			})
			os.Exit(handle.ExitError)
		}
	}()
	// 优先级：KUAKE_COOKIE（整段）> KUAKE_PUS + KUAKE_PUUS 拼接 > -cookies/--cookies
	if norm := sdk.ResolveEnvCookieString(); norm != "" {
		client = sdk.NewQuarkClient(norm)
	} else if cookies != "" {
		cookies = normalizeQuarkCookieInput(cookies)
		if cookies == "" {
			client = sdk.NewQuarkClient()
		} else {
			client = sdk.NewQuarkClient(cookies)
		}
	} else {
		client = sdk.NewQuarkClient()
	}

	// 执行命令
	var result *handle.CLIResult
	switch command {
	case "user":
		result = handle.UserInfo(client)
	case "list":
		result = handle.List(client, args)
	case "info":
		result = handle.Info(client, args)
	case "download":
		result = handle.Download(client, args)
	case "upload":
		result = handle.Upload(client, args)
	case "create":
		result = handle.CreateFolder(client, args)
	case "move":
		result = handle.Move(client, args)
	case "copy":
		result = handle.Copy(client, args)
	case "rename":
		result = handle.Rename(client, args)
	case "delete":
		result = handle.Delete(client, args)
	case "share":
		result = handle.ShareCreate(client, args)
	case "share-delete":
		result = handle.ShareDelete(client, args)
	case "share-list":
		result = handle.ShareList(client, args)
	case "share-save":
		result = handle.ShareSave(client, args)
	case "help", "-h", "--help":
		printUsage()
		os.Exit(handle.ExitSuccess)
	case "version", "-v", "--version":
		handle.OutputJSON(&handle.CLIResult{
			Success: true,
			Code:    "OK",
			Message: fmt.Sprintf("kuake %s", Version),
			Data: map[string]interface{}{
				"version": Version,
			},
		})
		os.Exit(handle.ExitSuccess)
	default:
		result = &handle.CLIResult{
			Success: false,
			Code:    "UNKNOWN_COMMAND",
			Message: fmt.Sprintf("Unknown command: %s", command),
		}
	}

	// 处理流式模式（result 为 nil 表示已经输出完毕）
	if result == nil {
		os.Exit(handle.ExitSuccess)
	}

	// 输出 JSON 结果
	handle.OutputJSON(result)

	// 根据结果设置退出码
	if !result.Success {
		os.Exit(handle.ExitError)
	}
	os.Exit(handle.ExitSuccess)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Quark Cloud Drive CLI Tool

Usage:
  kuake [options] <command> [arguments...]

Options:
  -cookies, --cookies <value>  Specify cookie value directly (only when KUAKE_COOKIE empty after trim;
                                adds __pus= prefix)
  -v, --version                Show version information

Auth:
  Env cookie: full KUAKE_COOKIE (after trim+normalize) OR split KUAKE_PUS + KUAKE_PUUS (values only, no __pus=/__puus= prefix), then -cookies/--cookies.
  BREAKING; see CHANGELOG.
  Optional .env: if .env exists in cwd, load it before creating the client (does not override existing env vars). Set KUAKE_LOAD_DOTENV=0 to disable.

Commands:
  user                        Get user information
  list [path] [--stream]     List directory (default: "/")
                              Use --stream to output one JSON per line for pipeline mode
  info <path>                 Get file/folder info (supports pipe mode)
  download <path> [dest]      Get file download URL, or download to local file if dest given (supports pipe mode)
  upload <file> <dest> [--max_upload_parallel N]
                              Upload file (all parameters must be quoted)
  create <name> <pdir>        Create folder (use "/" for root)
  move <src> <dest>           Move file/folder
  copy <src> <dest>           Copy file/folder
  rename <path> <newName>     Rename file/folder
  delete <path>               Delete file/folder (supports pipe mode)
  share <path> <days> <passcode>  Create share link
                                days: 0=permanent, 1/7/30=days
                                passcode: "true" or "false"
  share-delete <share_id_or_path>...  Delete share(s) by share ID(s) or file path(s)
  share-list [page] [size] [orderField] [orderType]  Get my share list
                                page: page number (default: 1)
                                size: page size (default: 50)
                                orderField: sort field (default: "created_at")
                                orderType: "asc" or "desc" (default: "desc")
  share-save <share_link> [passcode] [dest_dir]  Save shared files to your drive
                                share_link: share link (e.g., "https://pan.quark.cn/s/xxx")
                                passcode: extraction code (optional, auto-extracted from link if present)
                                dest_dir: destination directory (default: "/")
  version                     Show version information
  help                           Show help

Examples:
  kuake user
  kuake list "/"
  kuake info "/file.txt"
  kuake download "/file.txt"
  kuake download "/file.txt" .
  kuake download "/file.txt" ./local.zip
  kuake upload "file.txt" "/folder/file.txt"
  kuake upload "file.txt" "/folder/file.txt" --max_upload_parallel 4
  kuake create "folder" "/"
  kuake move "/file.txt" "/folder/"
  kuake share "/file.txt" 7 "false"
  kuake share-delete "fdd8bfd93f21491ab80122538bec310d"
  kuake share-delete "/file.txt"
  kuake share-list
  kuake share-list 1 50 "created_at" "desc"
  kuake share-save "https://pan.quark.cn/s/xxx"
  kuake share-save "https://pan.quark.cn/s/xxx" "1234" "/folder"
  
  # Using -cookies parameter:
  kuake -cookies "your_cookie_value_here" user
  kuake -cookies "your_cookie_value_here" upload "file.txt" "/folder/file.txt"

Pipeline Mode:
  Commands can be chained using Unix pipes. When stdin has data, commands automatically
  switch to pipe mode and process input line by line.
  
  Examples:
    # List files and delete them
    kuake list "/photos" --stream | kuake delete
    
    # List files and get info for each
    kuake list "/" --stream | kuake info
    
    # List files and get download URLs
    kuake list "/documents" --stream | kuake download
    
    # List files, filter with jq, then delete
    kuake list "/" --stream | jq -r 'select(.size > 1000000) | .path' | kuake delete

Notes:
  - All path parameters must be quoted
  - Root directory is "/"
  - Upload parallel: --max_upload_parallel overrides KUAKE_UPLOAD_PARALLEL when both apply; if only
    env is set (1-16) it is used when the flag is omitted (default 4)
  - Results output as JSON to stdout
  - Exit code: 0=success, 1=failure
  - If KUAKE_COOKIE is set (non-empty after trim), it wins over -cookies
  - In pipe mode, each input line should be a JSON object with "path" or "fid" field
  - Use --stream with list command to output one JSON per line for pipeline processing
`)
}
