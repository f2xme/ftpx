package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/f2xme/ftpx/pkg/util"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPClient SFTP 客户端实现
type SFTPClient struct {
	config     *Config
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	workDir    string
}

// NewSFTPClient 创建 SFTP 客户端
func NewSFTPClient(config *Config) (*SFTPClient, error) {
	if config.Protocol != ProtocolSFTP {
		return nil, fmt.Errorf("协议类型错误: 期望 SFTP，实际 %s", config.Protocol)
	}

	return &SFTPClient{
		config: config,
	}, nil
}

// Connect 连接到 SFTP 服务器
func (c *SFTPClient) Connect(ctx context.Context) error {
	// 构建 SSH 配置
	sshConfig := &ssh.ClientConfig{
		User:            c.config.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: 实现严格的主机密钥检查
		Timeout:         c.config.Options.Timeout,
	}

	// 配置认证方法
	switch c.config.Auth.Type {
	case AuthTypePassword:
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(c.config.Auth.Password),
		}

	case AuthTypeKey:
		key, err := c.loadPrivateKey(c.config.Auth.KeyFile, c.config.Auth.Passphrase)
		if err != nil {
			return fmt.Errorf("加载私钥失败: %w", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.PublicKeys(key),
		}

	case AuthTypeAgent:
		return fmt.Errorf("SSH Agent 认证暂未实现")

	default:
		return fmt.Errorf("不支持的认证类型: %s", c.config.Auth.Type)
	}

	// 连接到 SSH 服务器
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)
	sshClient, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %w", err)
	}

	// 创建 SFTP 会话
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return fmt.Errorf("SFTP 会话创建失败: %w", err)
	}

	c.sshClient = sshClient
	c.sftpClient = sftpClient

	// 获取当前工作目录
	if wd, err := c.sftpClient.Getwd(); err == nil {
		c.workDir = wd
	}

	return nil
}

// Close 关闭连接
func (c *SFTPClient) Close() error {
	var err error
	if c.sftpClient != nil {
		err = c.sftpClient.Close()
		c.sftpClient = nil
	}
	if c.sshClient != nil {
		if e := c.sshClient.Close(); e != nil && err == nil {
			err = e
		}
		c.sshClient = nil
	}
	return err
}

// IsConnected 检查是否已连接
func (c *SFTPClient) IsConnected() bool {
	return c.sftpClient != nil && c.sshClient != nil
}

// Upload 上传文件
func (c *SFTPClient) Upload(ctx context.Context, local, remote string, opts *TransferOptions) error {
	if opts == nil {
		opts = DefaultTransferOptions()
	}

	// 检查本地文件
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
func (c *SFTPClient) uploadFile(ctx context.Context, local, remote string, localInfo os.FileInfo, opts *TransferOptions) error {
	// 打开本地文件
	localFile, err := os.Open(local)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer localFile.Close()

	// 检查是否需要断点续传
	var offset int64
	if opts.Resume {
		if remoteInfo, err := c.sftpClient.Stat(remote); err == nil {
			offset = remoteInfo.Size()
			if offset >= localInfo.Size() {
				// 远程文件已完整
				return nil
			}
			localFile.Seek(offset, 0)
		}
	}

	// 创建远程文件
	var remoteFile *sftp.File
	if offset > 0 {
		remoteFile, err = c.sftpClient.OpenFile(remote, os.O_WRONLY|os.O_APPEND)
	} else {
		// 确保远程目录存在
		if err := c.ensureRemoteDir(filepath.Dir(remote)); err != nil {
			return err
		}
		remoteFile, err = c.sftpClient.Create(remote)
	}
	if err != nil {
		return fmt.Errorf("创建远程文件失败: %w", err)
	}
	defer remoteFile.Close()

	// 执行传输
	return c.copyWithProgress(ctx, remoteFile, localFile, offset, localInfo.Size(), opts)
}

// uploadDirectory 上传目录
func (c *SFTPClient) uploadDirectory(ctx context.Context, local, remote string, opts *TransferOptions) error {
	// 创建远程目录
	if err := c.Mkdir(ctx, remote, true); err != nil {
		return err
	}

	// 遍历本地目录
	return filepath.Walk(local, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查上下文取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 计算相对路径
		relPath, err := filepath.Rel(local, path)
		if err != nil {
			return err
		}
		remotePath := filepath.Join(remote, relPath)

		if info.IsDir() {
			// 对于目录，使用 MkdirAll 确保父目录存在
			return c.sftpClient.MkdirAll(remotePath)
		}

		return c.uploadFile(ctx, path, remotePath, info, opts)
	})
}

// Download 下载文件
func (c *SFTPClient) Download(ctx context.Context, remote, local string, opts *TransferOptions) error {
	if opts == nil {
		opts = DefaultTransferOptions()
	}

	// 检查远程文件
	remoteInfo, err := c.sftpClient.Stat(remote)
	if err != nil {
		return fmt.Errorf("远程文件不存在: %w", err)
	}

	if remoteInfo.IsDir() {
		return c.downloadDirectory(ctx, remote, local, opts)
	}

	return c.downloadFile(ctx, remote, local, remoteInfo, opts)
}

// downloadFile 下载单个文件
func (c *SFTPClient) downloadFile(ctx context.Context, remote, local string, remoteInfo os.FileInfo, opts *TransferOptions) error {
	// 确保本地目录存在
	if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	// 打开远程文件
	remoteFile, err := c.sftpClient.Open(remote)
	if err != nil {
		return fmt.Errorf("打开远程文件失败: %w", err)
	}
	defer remoteFile.Close()

	// 检查是否需要断点续传
	var offset int64
	var localFile *os.File
	if opts.Resume {
		if localInfo, err := os.Stat(local); err == nil {
			offset = localInfo.Size()
			if offset >= remoteInfo.Size() {
				// 本地文件已完整
				return nil
			}
			localFile, err = os.OpenFile(local, os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				return fmt.Errorf("打开本地文件失败: %w", err)
			}
			remoteFile.Seek(offset, 0)
		}
	}

	if localFile == nil {
		localFile, err = os.Create(local)
		if err != nil {
			return fmt.Errorf("创建本地文件失败: %w", err)
		}
	}
	defer localFile.Close()

	// 执行传输
	return c.copyWithProgress(ctx, localFile, remoteFile, offset, remoteInfo.Size(), opts)
}

// downloadDirectory 下载目录
func (c *SFTPClient) downloadDirectory(ctx context.Context, remote, local string, opts *TransferOptions) error {
	// 创建本地目录
	if err := os.MkdirAll(local, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	// 使用 Walk 遍历远程目录
	walker := c.sftpClient.Walk(remote)
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}

		// 检查上下文取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		path := walker.Path()
		info := walker.Stat()

		// 计算相对路径
		relPath, err := filepath.Rel(remote, path)
		if err != nil {
			return err
		}
		localPath := filepath.Join(local, relPath)

		if info.IsDir() {
			if err := os.MkdirAll(localPath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := c.downloadFile(ctx, path, localPath, info, opts); err != nil {
			return err
		}
	}

	return nil
}

// copyWithProgress 带进度的复制
func (c *SFTPClient) copyWithProgress(ctx context.Context, dst io.Writer, src io.Reader, offset, total int64, opts *TransferOptions) error {
	// 应用速率限制
	if opts.RateLimit > 0 {
		src = util.NewRateLimitedReader(src, opts.RateLimit)
	}

	buffer := make([]byte, opts.BufferSize)
	transferred := offset
	startTime := time.Now()

	for {
		// 检查上下文取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 读取数据
		n, err := src.Read(buffer)
		if n > 0 {
			// 写入数据
			if _, writeErr := dst.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}

			transferred += int64(n)

			// 调用进度回调
			if opts.OnProgress != nil {
				elapsed := time.Since(startTime).Seconds()
				speed := float64(transferred-offset) / elapsed
				opts.OnProgress(transferred, total, speed)
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// Remove 删除文件
func (c *SFTPClient) Remove(ctx context.Context, path string) error {
	return c.sftpClient.Remove(path)
}

// Rename 重命名文件
func (c *SFTPClient) Rename(ctx context.Context, oldPath, newPath string) error {
	return c.sftpClient.Rename(oldPath, newPath)
}

// Chmod 修改文件权限
func (c *SFTPClient) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return c.sftpClient.Chmod(path, mode)
}

// List 列出目录
func (c *SFTPClient) List(ctx context.Context, path string) ([]FileInfo, error) {
	entries, err := c.sftpClient.ReadDir(path)
	if err != nil {
		return nil, err
	}

	result := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		result = append(result, FileInfo{
			Name:    entry.Name(),
			Size:    entry.Size(),
			Mode:    entry.Mode(),
			ModTime: entry.ModTime(),
			IsDir:   entry.IsDir(),
		})
	}

	return result, nil
}

// Mkdir 创建目录
func (c *SFTPClient) Mkdir(ctx context.Context, path string, recursive bool) error {
	if recursive {
		return c.sftpClient.MkdirAll(path)
	}
	return c.sftpClient.Mkdir(path)
}

// RemoveDir 删除目录
func (c *SFTPClient) RemoveDir(ctx context.Context, path string, recursive bool) error {
	if !recursive {
		return c.sftpClient.RemoveDirectory(path)
	}

	// 递归删除
	walker := c.sftpClient.Walk(path)
	var dirs []string

	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}

		if walker.Stat().IsDir() {
			dirs = append(dirs, walker.Path())
		} else {
			if err := c.sftpClient.Remove(walker.Path()); err != nil {
				return err
			}
		}
	}

	// 反向删除目录
	for i := len(dirs) - 1; i >= 0; i-- {
		if err := c.sftpClient.RemoveDirectory(dirs[i]); err != nil {
			return err
		}
	}

	return nil
}

// Stat 获取文件信息
func (c *SFTPClient) Stat(ctx context.Context, path string) (FileInfo, error) {
	info, err := c.sftpClient.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// OpenReader 打开读取流
func (c *SFTPClient) OpenReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
	file, err := c.sftpClient.Open(path)
	if err != nil {
		return nil, err
	}

	if offset > 0 {
		if _, err := file.Seek(offset, 0); err != nil {
			file.Close()
			return nil, err
		}
	}

	return file, nil
}

// OpenWriter 打开写入流
func (c *SFTPClient) OpenWriter(ctx context.Context, path string, offset int64) (io.WriteCloser, error) {
	var file *sftp.File
	var err error

	if offset > 0 {
		file, err = c.sftpClient.OpenFile(path, os.O_WRONLY|os.O_APPEND)
	} else {
		file, err = c.sftpClient.Create(path)
	}

	return file, err
}

// WorkDir 获取当前工作目录
func (c *SFTPClient) WorkDir() (string, error) {
	if c.workDir != "" {
		return c.workDir, nil
	}
	return c.sftpClient.Getwd()
}

// ChangeDir 切换工作目录
func (c *SFTPClient) ChangeDir(path string) error {
	c.workDir = path
	return nil
}

// GetProtocol 获取协议类型
func (c *SFTPClient) GetProtocol() Protocol {
	return ProtocolSFTP
}

// loadPrivateKey 加载私钥
func (c *SFTPClient) loadPrivateKey(keyFile, passphrase string) (ssh.Signer, error) {
	keyData, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}

	var signer ssh.Signer
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyData)
	}

	return signer, err
}

// ensureRemoteDir 确保远程目录存在
func (c *SFTPClient) ensureRemoteDir(dir string) error {
	if dir == "" || dir == "." || dir == "/" {
		return nil
	}

	// 检查目录是否存在
	if _, err := c.sftpClient.Stat(dir); err == nil {
		return nil
	}

	// 递归创建父目录
	parent := filepath.Dir(dir)
	if err := c.ensureRemoteDir(parent); err != nil {
		return err
	}

	return c.sftpClient.Mkdir(dir)
}
