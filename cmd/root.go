package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	profile string
	verbose bool
)

// rootCmd 根命令
var rootCmd = &cobra.Command{
	Use:   "ftpx",
	Short: "功能全面的 FTP/SFTP 客户端 CLI 工具",
	Long: `ftpx 是一个统一的 FTP/SFTP 客户端命令行工具。

支持：
  • SFTP（密码/密钥认证）
  • FTP/FTPS（主动/被动模式）
  • 文件上传/下载（支持断点续传）
  • 目录同步（单向/双向）
  • 并发传输
  • 配置管理

示例：
  ftpx connect sftp://user@host:22
  ftpx upload -r ./local /remote
  ftpx sync push ./local /remote
  ftpx -p myprofile ls /remote/path`,
	Version: "0.1.0",
}

// Execute 执行根命令
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认 $HOME/.ftpx/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "使用指定的连接配置")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "详细输出")
}

// initConfig 初始化配置
func initConfig() {
	if cfgFile != "" {
		// 使用指定的配置文件
		viper.SetConfigFile(cfgFile)
	} else {
		// 查找默认配置文件
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "错误: 无法获取用户主目录:", err)
			os.Exit(1)
		}

		// 配置目录
		configDir := home + "/.ftpx"

		// 如果配置目录不存在，创建它
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				fmt.Fprintln(os.Stderr, "错误: 无法创建配置目录:", err)
				os.Exit(1)
			}
		}

		viper.AddConfigPath(configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()

	// 读取配置文件（如果存在）
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Println("使用配置文件:", viper.ConfigFileUsed())
		}
	}
}
