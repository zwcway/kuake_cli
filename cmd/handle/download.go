package handle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"github.com/zwcway/kuake_cli/sdk"
	"golang.org/x/sync/errgroup"
)

// Download 处理下载命令：download <path> [dest]
// 若提供 dest则下载到本地文件并输出进度；否则仅返回下载链接 JSON
func Download(client *sdk.QuarkClient, args []string) *CLIResult {
	// 检查是否有 stdin 输入（管道模式）
	destPath := ""
	if len(args) >= 1 {
		destPath = args[0] // 管道模式下，第一个参数可能是 dest
	}

	if hasStdinData() {
		processStdinLines(func(path, fid string) *CLIResult {
			// 优先使用 path，如果没有则使用 fid
			targetPath := path
			if targetPath == "" && fid != "" {
				// 只有 fid 时，尝试直接使用
				targetPath = fid
			}

			if targetPath == "" {
				return &CLIResult{
					Success: false,
					Code:    "INVALID_INPUT",
					Message: "cannot determine path or fid from input",
				}
			}

			return download(client, targetPath, destPath)
		})
		// processStdinLines 已经处理了所有输出，返回 nil 表示已完成
		return nil
	}

	// 普通模式：从命令行参数读取
	if len(args) < 1 {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_ARGS",
			Message: `Usage: download <path> [dest] (path must be quoted, e.g., download "/file.txt" or download "/file.txt" ./local) or use pipe mode`,
		}
	}

	path := args[0]
	destPath = ""
	if len(args) >= 2 {
		destPath = args[1]
	}

	return download(client, path, destPath)
}

func download(client *sdk.QuarkClient, path, destPath string) *CLIResult {
	if filepath.IsAbs(destPath) {
		destPath, _ = filepath.Abs(destPath)
	}

	destPath = filepath.Clean(destPath)

	fileInfo, err := client.GetFileInfo(path)
	if err != nil {
		return &CLIResult{
			Success: false,
			Message: fmt.Sprintf("failed to get file info: %v", err),
		}
	}
	if !fileInfo.Success {
		return &CLIResult{
			Success: false,
			Code:    fileInfo.Code,
			Message: fileInfo.Message,
		}
	}

	fid, ok := fileInfo.Data["fid"].(string)
	if !ok || fid == "" {
		return &CLIResult{
			Success: false,
			Code:    "INVALID_FILE_INFO",
			Message: "file info does not contain valid fid",
		}
	}

	isDir, _ := fileInfo.Data["dir"].(bool)
	if isDir {
		if destPath == "" {
			destPath, _ = os.Getwd()
		}
		var wg errgroup.Group
		wg.SetLimit(3)
		p := mpb.New(mpb.WithWidth(64))

		return downloadDir(client, &wg, p, path, destPath)
	}

	fileName, _ := fileInfo.Data["file_name"].(string)

	if fileName == "" {
		fileName = filepath.Base(path)
	}
	if fileName == "" || fileName == "." {
		fileName = "download"
	}

	// 指定了 dest：下载到本地
	if destPath != "" {
		fileSize, _ := fileInfo.Data["size"].(int64)

		return downloadFile(client, fid, path, destPath, fileName, fileSize)
	}

	// 未指定 dest：仅返回下载链接
	downloadURL, err := client.GetDownloadURL(fid)
	if err != nil {
		return &CLIResult{
			Success: false,
			Message: fmt.Sprintf("failed to get download URL: %v", err),
		}
	}
	return &CLIResult{
		Success: true,
		Code:    "OK",
		Message: "Download URL retrieved successfully",
		Data:    map[string]interface{}{"fid": fid, "path": path, "download_url": downloadURL},
	}
}

func downloadFile(client *sdk.QuarkClient, fid, path, destPath, fileName string, fileSize int64) *CLIResult {
	p := mpb.New(mpb.WithWidth(64))
	bar := p.AddBar(fileSize,
		mpb.PrependDecorators(decor.Counters(decor.SizeB1024(0), "Downloading %.1f / %.1f")),
		mpb.AppendDecorators(decor.Percentage()),
	)

	err := client.DownloadFile(fid, destPath, fileName, func(p *sdk.DownloadProgress) {
		bar.SetCurrent(p.Downloaded)
	})
	if err != nil {
		return &CLIResult{
			Success: false,
			Message: fmt.Sprintf("download failed: %v", err),
		}
	}
	// 解析最终本地路径（与 SDK 逻辑一致）
	localPath := destPath
	if destPath == "" || destPath == "." || strings.HasSuffix(destPath, "/") || strings.HasSuffix(destPath, string(filepath.Separator)) {
		localPath = filepath.Join(destPath, fileName)
	} else if info, err := os.Stat(destPath); err == nil && info.IsDir() {
		localPath = filepath.Join(destPath, fileName)
	}
	return &CLIResult{
		Success: true,
		Code:    "OK",
		Message: "File downloaded successfully",
		Data:    map[string]interface{}{"local_path": localPath, "path": path},
	}
}

func downloadDir(client *sdk.QuarkClient, wg *errgroup.Group, p *mpb.Progress, path, destPath string) *CLIResult {
	pathName := filepath.Base(path)
	destPath = filepath.Join(destPath, pathName)

	if s, err := os.Stat(destPath); os.IsNotExist(err) {
		if err = os.MkdirAll(destPath, 0755); err != nil {
			return &CLIResult{
				Success: false,
				Message: fmt.Sprintf("download dir failed: %v", err),
			}
		}
	} else if err != nil {
		return &CLIResult{
			Success: false,
			Message: fmt.Sprintf("download dir failed: %v", err),
		}
	} else if !s.IsDir() {
		return &CLIResult{
			Success: false,
			Message: "download dir failed: must special a directory",
		}
	}

	response, err := client.List(path)
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
	if quarkFileInfos, ok := response.Data["list"].([]sdk.QuarkFileInfo); ok {

		for i, qfi := range quarkFileInfos {

			if qfi.IsDirectory {
				if res := downloadDir(client, wg, p, qfi.Path, destPath); res != nil {
					return res
				}
				continue
			}

			wg.Go(func(nu int, qfi sdk.QuarkFileInfo) func() error {
				return func() error {

					dir := filepath.Dir(qfi.Path)
					fileName := filepath.Base(qfi.Name)

					relDir, err := filepath.Rel(path, dir)
					if err != nil {
						return err
					}
					dlDir := filepath.Join(destPath, relDir)
					relPath := filepath.Join(relDir, fileName)

					if _, err := os.Stat(dlDir); os.IsNotExist(err) {
						if err = os.MkdirAll(dlDir, 0755); err != nil {
							return err
						}
					} else if err != nil {
						return err
					}
					bar := p.AddBar(qfi.Size,
						mpb.PrependDecorators(decor.Name(fmt.Sprintf("[%2d/%2d] ", nu, len(quarkFileInfos)))),
						mpb.AppendDecorators(decor.Percentage(),
							decor.Counters(decor.SizeB1024(0), "%.1f/%.1f")),
						mpb.AppendDecorators(decor.Name(relPath)),
						mpb.BarWidth(15),
					)

					return client.DownloadFile(qfi.Fid, dlDir, fileName, func(p *sdk.DownloadProgress) { bar.SetCurrent(p.Downloaded) })
				}
			}(i+1, qfi))

		}
		err = wg.Wait()
		if err != nil {
			return &CLIResult{
				Success: false,
				Message: fmt.Sprintf("download failed: %v", err),
			}
		}
		return nil
	}

	return &CLIResult{}
}
