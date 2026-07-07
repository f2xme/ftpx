package config

import (
	"fmt"
	"os"
	"time"

	"github.com/f2xme/ftpx/pkg/client"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	Global   GlobalConfig          `yaml:"global"`
	Profiles map[string]*Profile   `yaml:"profiles"`
}

// GlobalConfig 全局设置
type GlobalConfig struct {
	LogLevel       string         `yaml:"log_level"`
	LogFile        string         `yaml:"log_file"`
	DefaultProfile string         `yaml:"default_profile"`
	Transfer       TransferConfig `yaml:"transfer"`
	Sync           SyncConfig     `yaml:"sync"`
}

// TransferConfig 传输设置
type TransferConfig struct {
	BufferSize    int           `yaml:"buffer_size"`
	ParallelCount int           `yaml:"parallel_count"`
	Timeout       time.Duration `yaml:"timeout"`
	RetryCount    int           `yaml:"retry_count"`
	RateLimit     int64         `yaml:"rate_limit"`
}

// SyncConfig 同步设置
type SyncConfig struct {
	ChecksumAlgorithm string   `yaml:"checksum_algorithm"`
	IgnorePatterns    []string `yaml:"ignore_patterns"`
}

// Profile 连接配置
type Profile struct {
	Protocol string        `yaml:"protocol"`
	Host     string        `yaml:"host"`
	Port     int           `yaml:"port"`
	User     string        `yaml:"user"`
	Auth     AuthConfig    `yaml:"auth"`
	Options  ClientOptions `yaml:"options"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type       string `yaml:"type"`
	Password   string `yaml:"password,omitempty"`
	KeyFile    string `yaml:"key_file,omitempty"`
	Passphrase string `yaml:"passphrase,omitempty"`
}

// ClientOptions 客户端选项
type ClientOptions struct {
	Compression bool          `yaml:"compression"`
	KeepAlive   time.Duration `yaml:"keep_alive"`
	TLSMode     string        `yaml:"tls_mode,omitempty"`
	PassiveMode bool          `yaml:"passive_mode,omitempty"`
}

// Load 加载配置
func Load() (*Config, error) {
	cfg := DefaultConfig()

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	return cfg, nil
}

// Save 保存配置
func Save(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := home + "/.ftpx/config.yaml"
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	return os.WriteFile(configPath, data, 0600)
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Global: GlobalConfig{
			LogLevel:       "info",
			LogFile:        "~/.ftpx/ftpx.log",
			DefaultProfile: "",
			Transfer: TransferConfig{
				BufferSize:    32768,
				ParallelCount: 3,
				Timeout:       30 * time.Second,
				RetryCount:    3,
				RateLimit:     0,
			},
			Sync: SyncConfig{
				ChecksumAlgorithm: "md5",
				IgnorePatterns: []string{
					".DS_Store",
					"Thumbs.db",
					".git/",
				},
			},
		},
		Profiles: make(map[string]*Profile),
	}
}

// GetProfile 获取指定的连接配置
func (c *Config) GetProfile(name string) (*Profile, error) {
	profile, ok := c.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("配置 '%s' 不存在", name)
	}
	return profile, nil
}

// AddProfile 添加连接配置
func (c *Config) AddProfile(name string, profile *Profile) error {
	if _, ok := c.Profiles[name]; ok {
		return fmt.Errorf("配置 '%s' 已存在", name)
	}
	c.Profiles[name] = profile
	return nil
}

// UpdateProfile 更新连接配置
func (c *Config) UpdateProfile(name string, profile *Profile) error {
	if _, ok := c.Profiles[name]; !ok {
		return fmt.Errorf("配置 '%s' 不存在", name)
	}
	c.Profiles[name] = profile
	return nil
}

// RemoveProfile 删除连接配置
func (c *Config) RemoveProfile(name string) error {
	if _, ok := c.Profiles[name]; !ok {
		return fmt.Errorf("配置 '%s' 不存在", name)
	}
	delete(c.Profiles, name)
	return nil
}

// ToClientConfig 转换为客户端配置
func (p *Profile) ToClientConfig() (*client.Config, error) {
	// 解析协议
	var protocol client.Protocol
	switch p.Protocol {
	case "sftp":
		protocol = client.ProtocolSFTP
	case "ftp":
		protocol = client.ProtocolFTP
	case "ftps":
		protocol = client.ProtocolFTPS
	default:
		return nil, fmt.Errorf("不支持的协议: %s", p.Protocol)
	}

	// 解析认证类型
	var authType client.AuthType
	switch p.Auth.Type {
	case "password":
		authType = client.AuthTypePassword
	case "key":
		authType = client.AuthTypeKey
	case "agent":
		authType = client.AuthTypeAgent
	default:
		return nil, fmt.Errorf("不支持的认证类型: %s", p.Auth.Type)
	}

	return &client.Config{
		Protocol: protocol,
		Host:     p.Host,
		Port:     p.Port,
		User:     p.User,
		Auth: client.AuthConfig{
			Type:       authType,
			Password:   p.Auth.Password,
			KeyFile:    p.Auth.KeyFile,
			Passphrase: p.Auth.Passphrase,
		},
		Options: client.ClientOptions{
			Compression: p.Options.Compression,
			KeepAlive:   p.Options.KeepAlive,
			Timeout:     30 * time.Second,
			TLSMode:     p.Options.TLSMode,
			PassiveMode: p.Options.PassiveMode,
		},
	}, nil
}
