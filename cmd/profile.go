package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/f2xme/ftpx/pkg/config"
	"github.com/spf13/cobra"
)

// profileCmd 配置管理命令
var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "管理连接配置",
	Long: `管理 FTP/SFTP 连接配置。

子命令：
  add     添加新的连接配置
  list    列出所有配置
  remove  删除配置
  show    显示配置详情`,
}

// profileAddCmd 添加配置命令
var profileAddCmd = &cobra.Command{
	Use:   "add <名称>",
	Short: "添加新的连接配置",
	Long: `添加新的连接配置。

示例：
  ftpx profile add myserver \
    --protocol sftp \
    --host example.com \
    --port 22 \
    --user admin \
    --auth-type key \
    --key-file ~/.ssh/id_rsa`,
	Args: cobra.ExactArgs(1),
	RunE: runProfileAdd,
}

// profileListCmd 列出配置命令
var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有连接配置",
	RunE:  runProfileList,
}

// profileRemoveCmd 删除配置命令
var profileRemoveCmd = &cobra.Command{
	Use:   "remove <名称>",
	Short: "删除连接配置",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileRemove,
}

// profileShowCmd 显示配置命令
var profileShowCmd = &cobra.Command{
	Use:   "show <名称>",
	Short: "显示配置详情",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileShow,
}

var (
	addProtocol    string
	addHost        string
	addPort        int
	addUser        string
	addAuthType    string
	addPassword    string
	addKeyFile     string
	addPassphrase  string
	addCompression bool
	addPassiveMode bool
)

func init() {
	rootCmd.AddCommand(profileCmd)
	profileCmd.AddCommand(profileAddCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileRemoveCmd)
	profileCmd.AddCommand(profileShowCmd)

	// add 命令的标志
	profileAddCmd.Flags().StringVar(&addProtocol, "protocol", "sftp", "协议类型 (sftp/ftp/ftps)")
	profileAddCmd.Flags().StringVar(&addHost, "host", "", "主机地址")
	profileAddCmd.Flags().IntVar(&addPort, "port", 22, "端口")
	profileAddCmd.Flags().StringVar(&addUser, "user", "", "用户名")
	profileAddCmd.Flags().StringVar(&addAuthType, "auth-type", "password", "认证类型 (password/key/agent)")
	profileAddCmd.Flags().StringVar(&addPassword, "password", "", "密码")
	profileAddCmd.Flags().StringVar(&addKeyFile, "key-file", "", "私钥文件路径")
	profileAddCmd.Flags().StringVar(&addPassphrase, "passphrase", "", "私钥密码")
	profileAddCmd.Flags().BoolVar(&addCompression, "compression", false, "启用压缩")
	profileAddCmd.Flags().BoolVar(&addPassiveMode, "passive", true, "FTP 被动模式")

	profileAddCmd.MarkFlagRequired("host")
	profileAddCmd.MarkFlagRequired("user")
}

func runProfileAdd(cmd *cobra.Command, args []string) error {
	name := args[0]

	// 加载现有配置
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// 扩展密钥文件路径中的 ~
	if addKeyFile != "" {
		if strings.HasPrefix(addKeyFile, "~/") {
			home, _ := os.UserHomeDir()
			addKeyFile = filepath.Join(home, addKeyFile[2:])
		}
	}

	// 创建新的配置
	prof := &config.Profile{
		Protocol: addProtocol,
		Host:     addHost,
		Port:     addPort,
		User:     addUser,
		Auth: config.AuthConfig{
			Type:       addAuthType,
			Password:   addPassword,
			KeyFile:    addKeyFile,
			Passphrase: addPassphrase,
		},
		Options: config.ClientOptions{
			Compression: addCompression,
			PassiveMode: addPassiveMode,
		},
	}

	// 添加到配置
	if err := cfg.AddProfile(name, prof); err != nil {
		return err
	}

	// 保存配置
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Printf("配置 '%s' 已添加\n", name)
	return nil
}

func runProfileList(cmd *cobra.Command, args []string) error {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if len(cfg.Profiles) == 0 {
		fmt.Println("没有配置")
		return nil
	}

	fmt.Println("可用的连接配置：")
	for name, prof := range cfg.Profiles {
		defaultMark := ""
		if name == cfg.Global.DefaultProfile {
			defaultMark = " (默认)"
		}
		fmt.Printf("  %s - %s://%s@%s:%d%s\n",
			name,
			prof.Protocol,
			prof.User,
			prof.Host,
			prof.Port,
			defaultMark,
		)
	}

	return nil
}

func runProfileRemove(cmd *cobra.Command, args []string) error {
	name := args[0]

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 删除配置
	if err := cfg.RemoveProfile(name); err != nil {
		return err
	}

	// 如果删除的是默认配置，清空默认配置
	if cfg.Global.DefaultProfile == name {
		cfg.Global.DefaultProfile = ""
	}

	// 保存配置
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Printf("配置 '%s' 已删除\n", name)
	return nil
}

func runProfileShow(cmd *cobra.Command, args []string) error {
	name := args[0]

	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// 获取配置
	prof, err := cfg.GetProfile(name)
	if err != nil {
		return err
	}

	// 显示配置详情
	fmt.Printf("配置名称: %s\n", name)
	fmt.Printf("协议:     %s\n", prof.Protocol)
	fmt.Printf("主机:     %s\n", prof.Host)
	fmt.Printf("端口:     %d\n", prof.Port)
	fmt.Printf("用户:     %s\n", prof.User)
	fmt.Printf("认证:     %s\n", prof.Auth.Type)

	if prof.Auth.KeyFile != "" {
		fmt.Printf("密钥文件: %s\n", prof.Auth.KeyFile)
	}

	fmt.Printf("压缩:     %v\n", prof.Options.Compression)

	if prof.Protocol == "ftp" || prof.Protocol == "ftps" {
		fmt.Printf("被动模式: %v\n", prof.Options.PassiveMode)
	}

	return nil
}
