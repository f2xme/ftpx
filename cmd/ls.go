package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	lsLong bool
	lsHuman bool
)

// lsCmd 列表命令
var lsCmd = &cobra.Command{
	Use:   "ls [远程路径]",
	Short: "列出远程目录内容",
	Long: `列出远程目录的文件和子目录。

示例：
  ftpx ls /remote/path
  ftpx ls -l /remote/path              # 详细信息
  ftpx ls -l --human-readable /remote/path # 人类可读的文件大小`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLs,
}

func init() {
	rootCmd.AddCommand(lsCmd)

	lsCmd.Flags().BoolVarP(&lsLong, "long", "l", false, "显示详细信息")
	lsCmd.Flags().BoolVar(&lsHuman, "human-readable", false, "以人类可读格式显示文件大小")
}

func runLs(cmd *cobra.Command, args []string) error {
	remotePath := "."
	if len(args) > 0 {
		remotePath = args[0]
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

	if verbose {
		fmt.Printf("正在连接到 %s...\n", profile)
	}
	if err := cli.Connect(ctx); err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}

	// 列出目录
	files, err := cli.List(context.Background(), remotePath)
	if err != nil {
		return fmt.Errorf("列表失败: %w", err)
	}

	// 输出结果
	if lsLong {
		for _, file := range files {
			var size string
			if lsHuman {
				size = humanizeBytes(file.Size)
			} else {
				size = fmt.Sprintf("%d", file.Size)
			}

			modTime := file.ModTime.Format("2006-01-02 15:04")
			fileType := "-"
			if file.IsDir {
				fileType = "d"
			}

			fmt.Printf("%s%s  %10s  %s  %s\n",
				fileType,
				file.Mode.String()[1:],
				size,
				modTime,
				file.Name,
			)
		}
	} else {
		for _, file := range files {
			if file.IsDir {
				fmt.Printf("%s/\n", file.Name)
			} else {
				fmt.Println(file.Name)
			}
		}
	}

	return nil
}

// humanizeBytes 将字节数转换为人类可读格式
func humanizeBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f%s", float64(bytes)/float64(div), units[exp])
}
