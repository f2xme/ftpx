package client

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/f2xme/ftpx/pkg/util"
	"github.com/jlaffaye/ftp"
)

// FTPClient FTP 客户端实现
type FTPClient struct {
	config     *Config
	ftpClient  *ftp.ServerConn
	workDir    string
}

// NewFTPClient 创建 FTP 客户端
func NewFTPClient(config *Config) (*FTPClient, error) {
	if config.Protocol != ProtocolFTP {
		return nil, fmt.Errorf("协议类型错误: 期望 FTP，实际 %s", config.Protocol)
	}

	return &FTPClient{
		config: config,
	}, nil
}

// Connect 连接到 FTP 服务器
func (c *FTPClient) Connect(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", c.config.Host, c.config.Port)

	// 创建连接选项
	options := []ftp.DialOption{
		ftp.DialWithTimeout(c.config.Options.Timeout),
	}

	// 连接到服务器
	conn, err := ftp.Dial(addr, options...)
	if err != nil {
		return fmt.Errorf("FTP 连接失败: %w", err)
	}

	// 登录
	if err := conn.Login(c.config.User, c.config.Auth.Password); err != nil {
		conn.Quit()
		return fmt.Errorf("FTP 登录失败: %w", err)
	}

	c.ftpClient = conn

	// 获取当前工作目录
	if wd, err := c.ftpClient.CurrentDir(); err == nil {
		c.workDir = wd
	}

	return nil
}

// Close 关闭连接
func (c *FTPClient) Close() error {
	if c.ftpClient != nil {
		err := c.ftpClient.Quit()
		c.ftpClient = nil
		return err
	}
	return nil
}

// IsConnected 检查是否已连接
func (c *FTPClient) IsConnected() bool {
	if c.ftpClient == nil {
		return false
	}
	// 尝试 NOOP 命令检查连接状态
	return c.ftpClient.NoOp() == nil
}

// Upload 上传文件
func (c *FTPClient) Upload(ctx context.Context, local, remote string, opts *TransferOptions) error {
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
func (c *FTPClient) uploadFile(ctx context.Context, local, remote string, localInfo os.FileInfo, opts *TransferOptions) error {
	// 打开本地文件
	localFile, err := os.Open(local)
	if err != nil {
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer localFile.Close()

	// 检查是否需要断点续传
	var offset int64
	if opts.Resume {
		if remoteInfo, err := c.ftpClient.GetEntry(remote); err == nil {
			offset = int64(remoteInfo.Size)
			if offset >= localInfo.Size() {
				// 远程文件已完整
				return nil
			}
			localFile.Seek(offset, 0)
		}
	}

	// 确保远程目录存在
	remoteDir := filepath.Dir(remote)
	if err := c.ensureRemoteDir(remoteDir); err != nil {
		return err
	}

	// 应用速率限制
	var reader io.Reader = localFile
	if opts.RateLimit > 0 {
		reader = util.NewRateLimitedReader(reader, opts.RateLimit)
	}

	// 创建带进度的 reader
	progressReader := &progressReader{
		reader:      reader,
		total:       localInfo.Size(),
		transferred: offset,
		onProgress:  opts.OnProgress,
		startTime:   time.Now(),
	}

	// 上传文件
	if err := c.ftpClient.Stor(remote, progressReader); err != nil {
		return fmt.Errorf("上传文件失败: %w", err)
	}

	return nil
}

// uploadDirectory 上传目录
func (c *FTPClient) uploadDirectory(ctx context.Context, local, remote string, opts *TransferOptions) error {
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
			// 递归创建目录，确保父目录存在
			return c.Mkdir(ctx, remotePath, true)
		}

		return c.uploadFile(ctx, path, remotePath, info, opts)
	})
}

// Download 下载文件
func (c *FTPClient) Download(ctx context.Context, remote, local string, opts *TransferOptions) error {
	if opts == nil {
		opts = DefaultTransferOptions()
	}

	// 检查远程文件 - 首先尝试 GetEntry
	remoteEntry, err := c.ftpClient.GetEntry(remote)
	if err != nil {
		// 如果 GetEntry 失败，尝试使用 List 作为后备
		dir := filepath.Dir(remote)
		name := filepath.Base(remote)
		entries, listErr := c.ftpClient.List(dir)
		if listErr != nil {
			return fmt.Errorf("远程文件不存在: %w", err)
		}

		// 在列表中查找文件
		found := false
		for _, e := range entries {
			if e.Name == name {
				remoteEntry = e
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("远程文件不存在: %s", remote)
		}
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
func (c *FTPClient) downloadFile(ctx context.Context, remote, local string, remoteInfo FileInfo, opts *TransferOptions) error {
	// 确保本地目录存在
	if err := os.MkdirAll(filepath.Dir(local), 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	// 检查是否需要断点续传
	var offset int64
	var localFile *os.File
	var err error

	if opts.Resume {
		if localInfo, err := os.Stat(local); err == nil {
			offset = localInfo.Size()
			if offset >= remoteInfo.Size {
				// 本地文件已完整
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

	// 下载文件
	response, err := c.ftpClient.Retr(remote)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer response.Close()

	// 如果需要续传，跳过已下载的部分
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

	// 创建带进度的 writer
	writer := &progressWriter{
		writer:      localFile,
		total:       remoteInfo.Size,
		transferred: offset,
		onProgress:  opts.OnProgress,
		startTime:   time.Now(),
	}

	// 执行下载
	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// downloadDirectory 下载目录
func (c *FTPClient) downloadDirectory(ctx context.Context, remote, local string, opts *TransferOptions) error {
	// 创建本地目录
	if err := os.MkdirAll(local, 0755); err != nil {
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	// 列出远程目录
	entries, err := c.ftpClient.List(remote)
	if err != nil {
		return fmt.Errorf("列出远程目录失败: %w", err)
	}

	for _, entry := range entries {
		// 检查上下文取消
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
func (c *FTPClient) Remove(ctx context.Context, path string) error {
	return c.ftpClient.Delete(path)
}

// Rename 重命名文件
func (c *FTPClient) Rename(ctx context.Context, oldPath, newPath string) error {
	return c.ftpClient.Rename(oldPath, newPath)
}

// Chmod 修改文件权限（FTP 不支持）
func (c *FTPClient) Chmod(ctx context.Context, path string, mode os.FileMode) error {
	return fmt.Errorf("FTP 协议不支持修改文件权限")
}

// List 列出目录
func (c *FTPClient) List(ctx context.Context, path string) ([]FileInfo, error) {
	entries, err := c.ftpClient.List(path)
	if err != nil {
		return nil, err
	}

	result := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		result = append(result, FileInfo{
			Name:    entry.Name,
			Size:    int64(entry.Size),
			Mode:    0644, // FTP 没有详细权限信息
			ModTime: entry.Time,
			IsDir:   entry.Type == ftp.EntryTypeFolder,
		})
	}

	return result, nil
}

// Mkdir 创建目录
func (c *FTPClient) Mkdir(ctx context.Context, path string, recursive bool) error {
	if !recursive {
		return c.ftpClient.MakeDir(path)
	}

	// 递归创建目录
	return c.ensureRemoteDir(path)
}

// RemoveDir 删除目录
func (c *FTPClient) RemoveDir(ctx context.Context, path string, recursive bool) error {
	if !recursive {
		return c.ftpClient.RemoveDir(path)
	}

	// 递归删除
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
func (c *FTPClient) Stat(ctx context.Context, path string) (FileInfo, error) {
	// 首先尝试 GetEntry
	entry, err := c.ftpClient.GetEntry(path)
	if err == nil {
		return FileInfo{
			Name:    entry.Name,
			Size:    int64(entry.Size),
			Mode:    0644,
			ModTime: entry.Time,
			IsDir:   entry.Type == ftp.EntryTypeFolder,
		}, nil
	}

	// 如果 GetEntry 失败，尝试列出父目录并查找文件
	// 这是一个后备方案，适用于不支持某些命令的 FTP 服务器
	dir := filepath.Dir(path)
	name := filepath.Base(path)

	entries, listErr := c.ftpClient.List(dir)
	if listErr != nil {
		// 如果列表也失败，返回原始错误
		return FileInfo{}, err
	}

	// 在列表中查找文件
	for _, e := range entries {
		if e.Name == name {
			return FileInfo{
				Name:    e.Name,
				Size:    int64(e.Size),
				Mode:    0644,
				ModTime: e.Time,
				IsDir:   e.Type == ftp.EntryTypeFolder,
			}, nil
		}
	}

	// 文件不存在
	return FileInfo{}, fmt.Errorf("文件不存在: %s", path)
}

// OpenReader 打开读取流
func (c *FTPClient) OpenReader(ctx context.Context, path string, offset int64) (io.ReadCloser, error) {
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
func (c *FTPClient) OpenWriter(ctx context.Context, path string, offset int64) (io.WriteCloser, error) {
	// FTP 的 Stor 命令会覆盖文件，不支持追加模式的流式写入
	return nil, fmt.Errorf("FTP 客户端不支持流式写入")
}

// WorkDir 获取当前工作目录
func (c *FTPClient) WorkDir() (string, error) {
	if c.workDir != "" {
		return c.workDir, nil
	}
	return c.ftpClient.CurrentDir()
}

// ChangeDir 切换工作目录
func (c *FTPClient) ChangeDir(path string) error {
	if err := c.ftpClient.ChangeDir(path); err != nil {
		return err
	}
	c.workDir = path
	return nil
}

// GetProtocol 获取协议类型
func (c *FTPClient) GetProtocol() Protocol {
	return ProtocolFTP
}

// ensureRemoteDir 确保远程目录存在
func (c *FTPClient) ensureRemoteDir(dir string) error {
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

// progressReader 带进度的 Reader
type progressReader struct {
	reader      io.Reader
	total       int64
	transferred int64
	onProgress  func(transferred, total int64, speed float64)
	startTime   time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.transferred += int64(n)
		if pr.onProgress != nil {
			elapsed := time.Since(pr.startTime).Seconds()
			speed := float64(pr.transferred) / elapsed
			pr.onProgress(pr.transferred, pr.total, speed)
		}
	}
	return n, err
}

// progressWriter 带进度的 Writer
type progressWriter struct {
	writer      io.Writer
	total       int64
	transferred int64
	onProgress  func(transferred, total int64, speed float64)
	startTime   time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if n > 0 {
		pw.transferred += int64(n)
		if pw.onProgress != nil {
			elapsed := time.Since(pw.startTime).Seconds()
			speed := float64(pw.transferred) / elapsed
			pw.onProgress(pw.transferred, pw.total, speed)
		}
	}
	return n, err
}
