package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/f2xme/ftpx/pkg/client"
	"github.com/f2xme/ftpx/pkg/config"
	"github.com/f2xme/ftpx/pkg/util"
	"github.com/spf13/cobra"
)

var (
	uploadRecursive         bool
	uploadResume            bool
	uploadOverwrite         bool
	uploadParallel          int
	uploadChecksum          bool
	uploadChecksumAlgorithm string
	uploadRateLimit         string
)

// uploadCmd 上传命令
var uploadCmd = &cobra.Command{
	Use:   "upload <本地路径> <远程路径>",
	Short: "上传文件或目录到远程服务器",
	Long: `上传文件或目录到远程服务器。

支持：
  • 单个文件上传
  • 递归目录上传 (-r)
  • 断点续传 (--resume)
  • 并发上传 (--parallel)

示例：
  ftpx upload file.txt /remote/path/
  ftpx upload -r ./dir /remote/dir
  ftpx upload --resume large-file.zip /remote/
  ftpx -p myprofile upload *.jpg /remote/images/`,
	Args: cobra.ExactArgs(2),
	RunE: runUpload,
}

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().BoolVarP(&uploadRecursive, "recursive", "r", false, "递归上传目录")
	uploadCmd.Flags().BoolVar(&uploadResume, "resume", false, "断点续传")
	uploadCmd.Flags().BoolVar(&uploadOverwrite, "overwrite", false, "覆盖已存在文件")
	uploadCmd.Flags().IntVar(&uploadParallel, "parallel", 3, "并发上传数")
	uploadCmd.Flags().BoolVar(&uploadChecksum, "checksum", false, "验证文件校验和")
	uploadCmd.Flags().StringVar(&uploadChecksumAlgorithm, "checksum-algorithm", "md5", "校验和算法 (md5, sha256)")
	uploadCmd.Flags().StringVar(&uploadRateLimit, "rate-limit", "", "速率限制 (例如: 1M, 500K, 10MB)")
}

func runUpload(cmd *cobra.Command, args []string) error {
	localPath := args[0]
	remotePath := args[1]

	// 检查本地文件是否存在
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("本地路径不存在: %w", err)
	}

	// 如果是目录但没有指定递归，提示错误
	if localInfo.IsDir() && !uploadRecursive {
		return fmt.Errorf("'%s' 是目录，请使用 -r 选项", localPath)
	}

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

	// 获取文件总大小（用于进度条）
	var totalSize int64
	if !localInfo.IsDir() {
		totalSize = localInfo.Size()
	}

	// 创建进度条
	var progressBar *util.ProgressBar
	if totalSize > 0 {
		progressBar = util.NewProgressBar(totalSize, "上传")
	}

	// 解析速率限制
	var rateLimit int64
	if uploadRateLimit != "" {
		var err error
		rateLimit, err = util.ParseRateLimit(uploadRateLimit)
		if err != nil {
			return fmt.Errorf("解析速率限制失败: %w", err)
		}
		if rateLimit > 0 && verbose {
			fmt.Printf("速率限制: %s\n", util.FormatSpeed(float64(rateLimit)))
		}
	}

	// 配置传输选项
	opts := &client.TransferOptions{
		Resume:            uploadResume,
		Overwrite:         uploadOverwrite,
		BufferSize:        32768,
		Parallel:          uploadParallel,
		Checksum:          uploadChecksum,
		ChecksumAlgorithm: uploadChecksumAlgorithm,
		RateLimit:         rateLimit,
		OnProgress: func(transferred, total int64, speed float64) {
			if progressBar != nil {
				progressBar.Update(transferred)
			} else {
				// 降级到简单进度显示
				percent := float64(transferred) / float64(total) * 100
				speedMB := speed / 1024 / 1024
				fmt.Printf("\r上传进度: %.1f%% (%.2f MB/s)  ", percent, speedMB)
			}
		},
	}

	// 执行上传
	ctx = context.Background()
	if progressBar == nil {
		fmt.Printf("正在上传 %s 到 %s...\n", localPath, remotePath)
	}

	startTime := time.Now()
	if err := cli.Upload(ctx, localPath, remotePath, opts); err != nil {
		if progressBar != nil {
			progressBar.Clear()
		}
		return fmt.Errorf("上传失败: %w", err)
	}

	if progressBar != nil {
		progressBar.Finish()
	} else {
		fmt.Println()
	}

	duration := time.Since(startTime)
	fmt.Printf("\n上传完成，耗时: %s\n", duration.Round(time.Millisecond))

	// 如果启用了校验和验证，计算并显示校验和
	if uploadChecksum {
		if verbose {
			fmt.Printf("正在计算 %s 校验和...\n", uploadChecksumAlgorithm)
		}
		checksum, err := util.CalculateFileChecksum(localPath, util.ChecksumAlgorithm(uploadChecksumAlgorithm))
		if err != nil {
			fmt.Printf("警告: 计算校验和失败: %v\n", err)
		} else {
			fmt.Printf("校验和 (%s): %s\n", uploadChecksumAlgorithm, checksum)
		}
	}

	return nil
}

// createClient 创建客户端
func createClient() (client.Client, error) {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		// 如果配置不存在，使用默认配置
		cfg = config.DefaultConfig()
	}

	// 获取 profile
	profileName := profile
	if profileName == "" {
		profileName = cfg.Global.DefaultProfile
	}
	if profileName == "" {
		return nil, fmt.Errorf("未指定连接配置，请使用 -p 选项或设置默认配置")
	}

	prof, err := cfg.GetProfile(profileName)
	if err != nil {
		return nil, err
	}

	// 转换为客户端配置
	clientConfig, err := prof.ToClientConfig()
	if err != nil {
		return nil, err
	}

	// 创建客户端
	switch clientConfig.Protocol {
	case client.ProtocolSFTP:
		return client.NewSFTPClient(clientConfig)
	case client.ProtocolFTP:
		return client.NewFTPClient(clientConfig)
	case client.ProtocolFTPS:
		return client.NewFTPSClient(clientConfig)
	default:
		return nil, fmt.Errorf("不支持的协议: %s", clientConfig.Protocol)
	}
}
