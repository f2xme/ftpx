package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/f2xme/ftpx/pkg/client"
	"github.com/f2xme/ftpx/pkg/util"
	"github.com/spf13/cobra"
)

var (
	downloadRecursive         bool
	downloadResume            bool
	downloadOverwrite         bool
	downloadChecksum          bool
	downloadChecksumAlgorithm string
	downloadRateLimit         string
)

// downloadCmd 下载命令
var downloadCmd = &cobra.Command{
	Use:   "download <远程路径> <本地路径>",
	Short: "从远程服务器下载文件或目录",
	Long: `从远程服务器下载文件或目录。

支持：
  • 单个文件下载
  • 递归目录下载 (-r)
  • 断点续传 (--resume)

示例：
  ftpx download /remote/file.txt ./local/
  ftpx download -r /remote/dir ./local/
  ftpx download --resume /remote/large.zip ./`,
	Args: cobra.ExactArgs(2),
	RunE: runDownload,
}

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().BoolVarP(&downloadRecursive, "recursive", "r", false, "递归下载目录")
	downloadCmd.Flags().BoolVar(&downloadResume, "resume", false, "断点续传")
	downloadCmd.Flags().BoolVar(&downloadOverwrite, "overwrite", false, "覆盖已存在文件")
	downloadCmd.Flags().BoolVar(&downloadChecksum, "checksum", false, "验证文件校验和")
	downloadCmd.Flags().StringVar(&downloadChecksumAlgorithm, "checksum-algorithm", "md5", "校验和算法 (md5, sha256)")
	downloadCmd.Flags().StringVar(&downloadRateLimit, "rate-limit", "", "速率限制 (例如: 1M, 500K, 10MB)")
}

func runDownload(cmd *cobra.Command, args []string) error {
	remotePath := args[0]
	localPath := args[1]

	// 创建客户端
	cli, err := createClient()
	if err != nil {
		return err
	}
	defer cli.Close()

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("正在连接到 %s...\n", profile)
	if err := cli.Connect(ctx); err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	fmt.Println("连接成功")

	// 检查远程文件
	remoteInfo, err := cli.Stat(context.Background(), remotePath)
	if err != nil {
		return fmt.Errorf("远程路径不存在: %w", err)
	}

	// 如果是目录但没有指定递归，提示错误
	if remoteInfo.IsDir && !downloadRecursive {
		return fmt.Errorf("'%s' 是目录，请使用 -r 选项", remotePath)
	}

	// 创建进度条
	var progressBar *util.ProgressBar
	if !remoteInfo.IsDir {
		progressBar = util.NewProgressBar(remoteInfo.Size, "下载")
	}

	// 解析速率限制
	var rateLimit int64
	if downloadRateLimit != "" {
		var err error
		rateLimit, err = util.ParseRateLimit(downloadRateLimit)
		if err != nil {
			return fmt.Errorf("解析速率限制失败: %w", err)
		}
		if rateLimit > 0 && verbose {
			fmt.Printf("速率限制: %s\n", util.FormatSpeed(float64(rateLimit)))
		}
	}

	// 配置传输选项
	opts := &client.TransferOptions{
		Resume:            downloadResume,
		Overwrite:         downloadOverwrite,
		BufferSize:        32768,
		Checksum:          downloadChecksum,
		ChecksumAlgorithm: downloadChecksumAlgorithm,
		RateLimit:         rateLimit,
		OnProgress: func(transferred, total int64, speed float64) {
			if progressBar != nil {
				progressBar.Update(transferred)
			} else {
				// 降级到简单进度显示
				percent := float64(transferred) / float64(total) * 100
				speedMB := speed / 1024 / 1024
				fmt.Printf("\r下载进度: %.1f%% (%.2f MB/s)  ", percent, speedMB)
			}
		},
	}

	// 执行下载
	ctx = context.Background()
	if progressBar == nil {
		fmt.Printf("正在下载 %s 到 %s...\n", remotePath, localPath)
	}

	startTime := time.Now()
	if err := cli.Download(ctx, remotePath, localPath, opts); err != nil {
		if progressBar != nil {
			progressBar.Clear()
		}
		return fmt.Errorf("下载失败: %w", err)
	}

	if progressBar != nil {
		progressBar.Finish()
	} else {
		fmt.Println()
	}

	duration := time.Since(startTime)
	fmt.Printf("\n下载完成，耗时: %s\n", duration.Round(time.Millisecond))

	// 如果启用了校验和验证，计算并显示校验和
	if downloadChecksum {
		if verbose {
			fmt.Printf("正在计算 %s 校验和...\n", downloadChecksumAlgorithm)
		}
		checksum, err := util.CalculateFileChecksum(localPath, util.ChecksumAlgorithm(downloadChecksumAlgorithm))
		if err != nil {
			fmt.Printf("警告: 计算校验和失败: %v\n", err)
		} else {
			fmt.Printf("校验和 (%s): %s\n", downloadChecksumAlgorithm, checksum)
		}
	}

	return nil
}
