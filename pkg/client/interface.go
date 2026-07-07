package client

import (
	"context"
	"io"
	"os"
	"time"
)

// Protocol 协议类型
type Protocol int

const (
	ProtocolSFTP Protocol = iota
	ProtocolFTP
	ProtocolFTPS
)

func (p Protocol) String() string {
	switch p {
	case ProtocolSFTP:
		return "SFTP"
	case ProtocolFTP:
		return "FTP"
	case ProtocolFTPS:
		return "FTPS"
	default:
		return "Unknown"
	}
}

// Client 定义统一的客户端接口
type Client interface {
	// 连接管理
	Connect(ctx context.Context) error
	Close() error
	IsConnected() bool

	// 文件操作
	Upload(ctx context.Context, local, remote string, opts *TransferOptions) error
	Download(ctx context.Context, remote, local string, opts *TransferOptions) error
	Remove(ctx context.Context, path string) error
	Rename(ctx context.Context, oldPath, newPath string) error
	Chmod(ctx context.Context, path string, mode os.FileMode) error

	// 目录操作
	List(ctx context.Context, path string) ([]FileInfo, error)
	Mkdir(ctx context.Context, path string, recursive bool) error
	RemoveDir(ctx context.Context, path string, recursive bool) error
	Stat(ctx context.Context, path string) (FileInfo, error)

	// 流式操作
	OpenReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error)
	OpenWriter(ctx context.Context, path string, offset int64) (io.WriteCloser, error)

	// 辅助方法
	WorkDir() (string, error)
	ChangeDir(path string) error
	GetProtocol() Protocol
}

// FileInfo 文件信息
type FileInfo struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
	Link    string // 符号链接目标
}

// TransferOptions 传输选项
type TransferOptions struct {
	Resume            bool                                            // 断点续传
	Overwrite         bool                                            // 覆盖已存在文件
	BufferSize        int                                             // 缓冲区大小
	RateLimit         int64                                           // 速率限制 (bytes/sec)
	OnProgress        func(transferred, total int64, speed float64)  // 进度回调
	Checksum          bool                                            // 校验和验证
	ChecksumAlgorithm string                                          // 校验和算法 (md5, sha256)
	Compress          bool                                            // 压缩传输
	Parallel          int                                             // 并发数（目录传输时）
}

// DefaultTransferOptions 返回默认传输选项
func DefaultTransferOptions() *TransferOptions {
	return &TransferOptions{
		Resume:            false,
		Overwrite:         false,
		BufferSize:        32768, // 32KB
		RateLimit:         0,     // 无限制
		OnProgress:        nil,
		Checksum:          false,
		ChecksumAlgorithm: "md5",
		Compress:          false,
		Parallel:          3,
	}
}

// Config 客户端配置
type Config struct {
	Protocol Protocol
	Host     string
	Port     int
	User     string
	Auth     AuthConfig
	Options  ClientOptions
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type       AuthType // password, key, agent
	Password   string
	KeyFile    string
	Passphrase string // 密钥密码
}

// AuthType 认证类型
type AuthType string

const (
	AuthTypePassword AuthType = "password"
	AuthTypeKey      AuthType = "key"
	AuthTypeAgent    AuthType = "agent"
)

// ClientOptions 客户端选项
type ClientOptions struct {
	Compression bool          // 是否启用压缩
	KeepAlive   time.Duration // 保活间隔
	Timeout     time.Duration // 连接超时
	TLSMode     string        // TLS 模式 (explicit, implicit)
	PassiveMode bool          // FTP 被动模式
}
