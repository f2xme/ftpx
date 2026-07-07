package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/f2xme/ftpx/pkg/util"
	"github.com/jlaffaye/ftp"
)

// FTPSClient FTPS 客户端实现
type FTPSClient struct {
	config     *Config
	ftpClient  *ftp.ServerConn
	workDir    string
}

// NewFTPSClient 创建 FTPS 客户端
func NewFTPSClient(config *Config) (*FTPSClient, error) {
	if config.Protocol != ProtocolFTPS {
		return nil, fmt.Errorf("协议类型错误: 期望 FTPS，实际 %s", config.Protocol)
	}

	return &FTPSClient{
		config: config,
	}, nil
}

// Connect 连接到 FTPS 服务器
func (c *FTPSClient) Connect(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// 配置 TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // TODO: 实现严格的证书验证
		ServerName:         c.config.Host,
	}

	// 创建连接选项
	options := []ftp.DialOption{
		ftp.DialWithTimeout(c.config.Options.Timeout),
	}

	// 根据 TLS 模式选择连接方式
	var conn *ftp.ServerConn
	var err error

	switch c.config.Options.TLSMode {
	case "explicit", "":
		// 显式 TLS（默认）- 先建立明文连接，然后升级到 TLS
		options = append(options, ftp.DialWithExplicitTLS(tlsConfig))
		conn, err = ftp.Dial(addr, options...)

	case "implicit":
		// 隐式 TLS - 从一开始就使用 TLS
		options = append(options, ftp.DialWithTLS(tlsConfig))
		conn, err = ftp.Dial(addr, options...)

	default:
		return fmt.Errorf("不支持的 TLS 模式: %s", c.config.Options.TLSMode)
	}

	if err != nil {
		return fmt.Errorf("FTPS 连接失败: %w", err)
	}

	// 登录
	if err := conn.Login(c.config.User, c.config.Auth.Password); err != nil {
		conn.Quit()
		return fmt.Errorf("FTPS 登录失败: %w", err)
	}

	c.ftpClient = conn

	// 获取当前工作目录
	if wd, err := c.ftpClient.CurrentDir(); err == nil {
		c.workDir = wd
	}

	return nil
}

// Close 关闭连接
func (c *FTPSClient) Close() error {
	if c.ftpClient != nil {
		err := c.ftpClient.Quit()
		c.ftpClient = nil
		return err
	}
	return nil
}

// IsConnected 检查是否已连接
func (c *FTPSClient) IsConnected() bool {
	if c.ftpClient == nil {
		return false
	}
	return c.ftpClient.NoOp() == nil
}

// Upload 上传文件
func (c *FTPSClient) Upload(ctx context.Context, local, remote string, opts *TransferOptions) error {
	if opts == nil {
		opts = DefaultTransferOptions()
	}

	localInfo, err := os.Stat(local)
	if err != nil {
		return fmt.Errorf("本地文件不存在: %w", err)
	}

	if localInfo.IsDir() {
		return c.uploadDirectory(ctx, local, remote, opts)
	}

	return c.uploadFile(ctx, local, remote, localInfo, opts)
}

// uploadFile 上传单个文件
func (c *FTPSClient) uploadFile(ctx context.Context, local, remote string, localInfo os.FileInfo, opts *TransferOptions) error {
	localFile, err := os.Open(local)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer localFile.Close()

	var offset int64
	if opts.Resume {
		if remoteInfo, err := c.ftpClient.GetEntry(remote); err == nil {
			offset = int64(remoteInfo.Size)
			if offset >= localInfo.Size() {
				return nil
			}
			localFile.Seek(offset, 0)
		}
	}

	remoteDir := filepath.Dir(remote)
	if err := c.ensureRemoteDir(remoteDir); err != nil {
		return err
	}

	// 应用速率限制
	var reader io.Reader = localFile
	if opts.RateLimit > 0 {
		reader = util.NewRateLimitedReader(reader, opts.RateLimit)
	}

	progressReader := &progressReader{
		reader:      reader,
		total:       localInfo.Size(),
		transferred: offset,
		onProgress:  opts.OnProgress,
		startTime:   time.Now(),
	}

	if err := c.ftpClient.Stor(remote, progressReader); err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	return nil
}

// uploadDirectory 上传目录
func (c *FTPSClient) uploadDirectory(ctx context.Context, local, remote string, opts *TransferOptions) error {
	if err := c.Mkdir(ctx, remote, true); err != nil {
		return err
	}

	return filepath.Walk(local, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		relPath, err := filepath.Rel(local, path)
		if err != nil {
			return err
		}
		remotePath := filepath.Join(remote, relPath)

		if info.IsDir() {
			// 递归创建目录，确保父目录存在
			return c.Mkdir(ctx, remotePath, true)
		}

		return c.uploadFile(ctx, path, remotePath, info, opts)
	})
}

// Download 下载文件
func (c *FTPSClient) Download(ctx context.Context, remote, local string, opts *TransferOptions) error {
	if opts == nil {
		opts = DefaultTransferOptions()
	}

	remoteEntry, err := c.ftpClient.GetEntry(remote)
	if err != nil {
		return fmt.Errorf("远程文件不存在: %w", err)
	}

	if remoteEntry.Type == ftp.EntryTypeFolder {
		return c.downloadDirectory(ctx, remote, local, opts)
	}

	remoteInfo := FileInfo{
		Name:    remoteEntry.Name,
		Size:    int64(remoteEntry.Size),
		ModTime: remoteEntry.Time,
		IsDir:   false,
	}

	return c.downloadFile(ctx, remote, local, remoteInfo, opts)
}

// downloadFile 下载单个文件
func (c *FTPSClient) downloadFile(ctx context.Context, remote, local string, remoteInfo FileInfo, opts *TransferOptions) error {
	if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	var offset int64
	var localFile *os.File
	var err error

	if opts.Resume {
		if localInfo, err := os.Stat(local); err == nil {
			offset = localInfo.Size()
			if offset >= remoteInfo.Size {
				return nil
			}
			localFile, err = os.OpenFile(local, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return fmt.Errorf("打开本地文件失败: %w", err)
			}
		}
	}

	if localFile == nil {
		localFile, err = os.Create(local)
		if err != nil {
			return fmt.Errorf("创建本地文件失败: %w", err)
		}
	}
	defer localFile.Close()

	response, err := c.ftpClient.Retr(remote)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer response.Close()

	if offset > 0 {
		_, err = io.CopyN(io.Discard, response, offset)
		if err != nil {
			return fmt.Errorf("跳到断点位置失败: %w", err)
		}
	}

	// 应用速率限制
	var reader io.Reader = response
	if opts.RateLimit > 0 {
		reader = util.NewRateLimitedReader(reader, opts.RateLimit)
	}

	writer := &progressWriter{
		writer:      localFile,
		total:       remoteInfo.Size,
		transferred: offset,
		onProgress:  opts.OnProgress,
		startTime:   time.Now(),
	}

	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// downloadDirectory 下载目录
func (c *FTPSClient) downloadDirectory(ctx context.Context, remote, local string, opts *TransferOptions) error {
	if err := os.MkdirAll(local, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	entries, err := c.ftpClient.List(remote)
	if err != nil {
		return fmt.Errorf("列出远程目录失败: %w", err)
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		remotePath := filepath.Join(remote, entry.Name)
		localPath := filepath.Join(local, entry.Name)

		if entry.Type == ftp.EntryTypeFolder {
			if err := c.downloadDirectory(ctx, remotePath, localPath, opts); err != nil {
				return err
			}
		} else {
			info := FileInfo{
				Name:    entry.Name,
				Size:    int64(entry.Size),
				ModTime: entry.Time,
				IsDir:   false,
			}
			if err := c.downloadFile(ctx, remotePath, localPath, info, opts); err != nil {
				return err
			}
		}
	}

	return nil
}

// Remove 删除文件
func (c *FTPSClient) Remove(ctx context.Context, path string) error {
	return c.ftpClient.Delete(path)
}

// Rename 重命名文件
func (c *FTPSClient) Rename(ctx context.Context, oldPath, newPath string) error {
	return c.ftpClient.Rename(oldPath, newPath)
}

// Chmod 修改文件权限
func (c *FTPSClient) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return fmt.Errorf("FTPS 协议不支持修改文件权限")
}

// List 列出目录
func (c *FTPSClient) List(ctx context.Context, path string) ([]FileInfo, error) {
	entries, err := c.ftpClient.List(path)
	if err != nil {
		return nil, err
	}

	result := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		result = append(result, FileInfo{
			Name:    entry.Name,
			Size:    int64(entry.Size),
			Mode:    0644,
			ModTime: entry.Time,
			IsDir:   entry.Type == ftp.EntryTypeFolder,
		})
	}

	return result, nil
}

// Mkdir 创建目录
func (c *FTPSClient) Mkdir(ctx context.Context, path string, recursive bool) error {
	if !recursive {
		return c.ftpClient.MakeDir(path)
	}
	return c.ensureRemoteDir(path)
}

// RemoveDir 删除目录
func (c *FTPSClient) RemoveDir(ctx context.Context, path string, recursive bool) error {
	if !recursive {
		return c.ftpClient.RemoveDir(path)
	}

	entries, err := c.ftpClient.List(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name)
		if entry.Type == ftp.EntryTypeFolder {
			if err := c.RemoveDir(ctx, entryPath, true); err != nil {
				return err
			}
		} else {
			if err := c.ftpClient.Delete(entryPath); err != nil {
				return err
			}
		}
	}

	return c.ftpClient.RemoveDir(path)
}

// Stat 获取文件信息
func (c *FTPSClient) Stat(ctx context.Context, path string) (FileInfo, error) {
	entry, err := c.ftpClient.GetEntry(path)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Name:    entry.Name,
		Size:    int64(entry.Size),
		Mode:    0644,
		ModTime: entry.Time,
		IsDir:   entry.Type == ftp.EntryTypeFolder,
	}, nil
}

// OpenReader 打开读取流
func (c *FTPSClient) OpenReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	response, err := c.ftpClient.Retr(path)
	if err != nil {
		return nil, err
	}

	if offset > 0 {
		_, err = io.CopyN(io.Discard, response, offset)
		if err != nil {
			response.Close()
			return nil, err
		}
	}

	return response, nil
}

// OpenWriter 打开写入流
func (c *FTPSClient) OpenWriter(ctx context.Context, path string, offset int64) (io.WriteCloser, error) {
	return nil, fmt.Errorf("FTPS 客户端不支持流式写入")
}

// WorkDir 获取当前工作目录
func (c *FTPSClient) WorkDir() (string, error) {
	if c.workDir != "" {
		return c.workDir, nil
	}
	return c.ftpClient.CurrentDir()
}

// ChangeDir 切换工作目录
func (c *FTPSClient) ChangeDir(path string) error {
	if err := c.ftpClient.ChangeDir(path); err != nil {
		return err
	}
	c.workDir = path
	return nil
}

// GetProtocol 获取协议类型
func (c *FTPSClient) GetProtocol() Protocol {
	return ProtocolFTPS
}

// ensureRemoteDir 确保远程目录存在
func (c *FTPSClient) ensureRemoteDir(dir string) error {
	if dir == "" || dir == "." || dir == "/" {
		return nil
	}

	// 检查目录是否存在
	_, err := c.ftpClient.GetEntry(dir)
	if err == nil {
		return nil // 目录已存在
	}

	// GetEntry 失败，尝试用 List 检查（后备机制）
	parent := filepath.Dir(dir)
	name := filepath.Base(dir)
	entries, listErr := c.ftpClient.List(parent)
	if listErr == nil {
		// 在列表中查找目录
		for _, e := range entries {
			if e.Name == name && e.Type == ftp.EntryTypeFolder {
				return nil // 目录存在
			}
		}
	}

	// 递归创建父目录
	if err := c.ensureRemoteDir(parent); err != nil {
		return err
	}

	// 尝试创建目录
	err = c.ftpClient.MakeDir(dir)
	if err != nil {
		// 再次检查目录是否已存在（可能是并发创建）
		if _, checkErr := c.ftpClient.GetEntry(dir); checkErr == nil {
			return nil
		}
	}
	return err
}
