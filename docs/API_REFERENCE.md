# FTPCli API 文档

本文档面向需要将 ftpx 作为 Go 库使用的开发者。

## 目录

- [安装](#安装)
- [核心接口](#核心接口)
- [客户端类型](#客户端类型)
- [配置管理](#配置管理)
- [传输选项](#传输选项)
- [错误处理](#错误处理)
- [使用示例](#使用示例)

## 安装

```bash
go get github.com/f2xme/ftpx
```

## 核心接口

### Client 接口

所有协议客户端（FTP、FTPS、SFTP）都实现了 `Client` 接口：

```go
type Client interface {
    // 连接管理
    Connect(ctx context.Context) error
    Disconnect() error
    IsConnected() bool
    
    // 目录操作
    List(ctx context.Context, path string) ([]FileInfo, error)
    Mkdir(ctx context.Context, path string, recursive bool) error
    Remove(ctx context.Context, path string, recursive bool) error
    Rename(ctx context.Context, oldPath, newPath string) error
    
    // 文件操作
    Stat(ctx context.Context, path string) (*FileInfo, error)
    Chmod(ctx context.Context, path string, mode os.FileMode) error
    
    // 传输操作
    Upload(ctx context.Context, local, remote string, opts *TransferOptions) error
    Download(ctx context.Context, remote, local string, opts *TransferOptions) error
    
    // 工作目录
    Pwd(ctx context.Context) (string, error)
    Cd(ctx context.Context, path string) error
}
```

### FileInfo 结构

```go
type FileInfo struct {
    Name    string
    Size    int64
    Mode    os.FileMode
    ModTime time.Time
    IsDir   bool
}
```

## 客户端类型

### FTP 客户端

```go
import "github.com/f2xme/ftpx/pkg/client"

// 创建配置
config := &client.Config{
    Protocol: client.ProtocolFTP,
    Host:     "ftp.example.com",
    Port:     21,
    User:     "username",
    Auth: client.AuthConfig{
        Type:     client.AuthTypePassword,
        Password: "password",
    },
    Options: client.Options{
        PassiveMode: true,
        Timeout:     30 * time.Second,
    },
}

// 创建客户端
ftpClient, err := client.NewFTPClient(config)
if err != nil {
    log.Fatal(err)
}
defer ftpClient.Disconnect()

// 连接
ctx := context.Background()
if err := ftpClient.Connect(ctx); err != nil {
    log.Fatal(err)
}
```

### FTPS 客户端

```go
config := &client.Config{
    Protocol: client.ProtocolFTPS,
    Host:     "ftps.example.com",
    Port:     990,
    User:     "username",
    Auth: client.AuthConfig{
        Type:     client.AuthTypePassword,
        Password: "password",
    },
    Options: client.Options{
        PassiveMode:    true,
        TLSMode:        "implicit", // 或 "explicit"
        TLSSkipVerify:  false,
        Timeout:        30 * time.Second,
    },
}

ftpsClient, err := client.NewFTPSClient(config)
if err != nil {
    log.Fatal(err)
}
```

### SFTP 客户端

#### 密码认证

```go
config := &client.Config{
    Protocol: client.ProtocolSFTP,
    Host:     "sftp.example.com",
    Port:     22,
    User:     "username",
    Auth: client.AuthConfig{
        Type:     client.AuthTypePassword,
        Password: "password",
    },
    Options: client.Options{
        Compression: false,
        Timeout:     30 * time.Second,
    },
}

sftpClient, err := client.NewSFTPClient(config)
if err != nil {
    log.Fatal(err)
}
```

#### 密钥认证

```go
config := &client.Config{
    Protocol: client.ProtocolSFTP,
    Host:     "sftp.example.com",
    Port:     22,
    User:     "username",
    Auth: client.AuthConfig{
        Type:        client.AuthTypeKey,
        KeyFile:     "/home/user/.ssh/id_rsa",
        KeyPassword: "",  // 如果密钥有密码
    },
    Options: client.Options{
        Compression: false,
        Timeout:     30 * time.Second,
    },
}

sftpClient, err := client.NewSFTPClient(config)
if err != nil {
    log.Fatal(err)
}
```

## 配置管理

### Config 结构

```go
type Config struct {
    Protocol Protocol
    Host     string
    Port     int
    User     string
    Auth     AuthConfig
    Options  Options
}

type Protocol string

const (
    ProtocolFTP  Protocol = "ftp"
    ProtocolFTPS Protocol = "ftps"
    ProtocolSFTP Protocol = "sftp"
)

type AuthConfig struct {
    Type        AuthType
    Password    string
    KeyFile     string
    KeyPassword string
}

type AuthType string

const (
    AuthTypePassword AuthType = "password"
    AuthTypeKey      AuthType = "key"
)

type Options struct {
    // 通用选项
    Compression bool
    KeepAlive   time.Duration
    Timeout     time.Duration
    
    // FTP/FTPS 专用
    PassiveMode   bool
    TLSMode       string  // "explicit" 或 "implicit"
    TLSSkipVerify bool
}
```

### 配置文件管理

```go
import "github.com/f2xme/ftpx/pkg/config"

// 加载配置文件
cfg, err := config.Load("/path/to/config.yaml")
if err != nil {
    log.Fatal(err)
}

// 获取配置
profile := cfg.GetProfile("myserver")
if profile == nil {
    log.Fatal("配置不存在")
}

// 添加配置
cfg.AddProfile("newserver", &config.Profile{
    Protocol: "ftp",
    Host:     "ftp.example.com",
    Port:     21,
    User:     "username",
    Auth: config.AuthConfig{
        Type:     "password",
        Password: "password",
    },
})

// 保存配置
if err := cfg.Save(); err != nil {
    log.Fatal(err)
}
```

## 传输选项

### TransferOptions 结构

```go
type TransferOptions struct {
    // 基本选项
    Overwrite bool
    Resume    bool
    Recursive bool
    
    // 性能选项
    RateLimit     int64  // 字节/秒，0 表示不限制
    ParallelCount int    // 并发数
    BufferSize    int    // 缓冲区大小
    
    // 校验选项
    Checksum          bool
    ChecksumAlgorithm string  // "md5" 或 "sha256"
    
    // 回调函数
    ProgressCallback func(progress TransferProgress)
}

type TransferProgress struct {
    BytesTransferred int64
    TotalBytes       int64
    Percentage       float64
    Speed            float64  // 字节/秒
    Elapsed          time.Duration
    Remaining        time.Duration
}
```

### 使用传输选项

```go
opts := &client.TransferOptions{
    Overwrite:         true,
    Resume:            true,
    RateLimit:         1024 * 1024,  // 1 MB/s
    Checksum:          true,
    ChecksumAlgorithm: "sha256",
    ProgressCallback: func(progress client.TransferProgress) {
        fmt.Printf("进度: %.2f%% (%d/%d bytes) %.2f KB/s\n",
            progress.Percentage,
            progress.BytesTransferred,
            progress.TotalBytes,
            progress.Speed/1024)
    },
}

err := ftpClient.Upload(ctx, "local.txt", "/remote/path.txt", opts)
if err != nil {
    log.Fatal(err)
}
```

## 错误处理

### 错误类型

```go
import "github.com/f2xme/ftpx/pkg/errors"

// 检查错误类型
if err != nil {
    switch {
    case errors.IsConnectionError(err):
        log.Println("连接错误:", err)
    case errors.IsAuthError(err):
        log.Println("认证失败:", err)
    case errors.IsNotFoundError(err):
        log.Println("文件不存在:", err)
    case errors.IsPermissionError(err):
        log.Println("权限拒绝:", err)
    default:
        log.Println("其他错误:", err)
    }
}
```

### 自定义错误

```go
type Error struct {
    Code    ErrorCode
    Message string
    Cause   error
}

type ErrorCode int

const (
    ErrConnection ErrorCode = iota + 1
    ErrAuthentication
    ErrNotFound
    ErrPermission
    ErrTimeout
    ErrTransfer
)

func (e *Error) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Cause)
    }
    return e.Message
}
```

## 使用示例

### 示例 1：基本文件上传

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/f2xme/ftpx/pkg/client"
)

func main() {
    config := &client.Config{
        Protocol: client.ProtocolFTP,
        Host:     "ftp.example.com",
        Port:     21,
        User:     "username",
        Auth: client.AuthConfig{
            Type:     client.AuthTypePassword,
            Password: "password",
        },
        Options: client.Options{
            PassiveMode: true,
            Timeout:     30 * time.Second,
        },
    }
    
    ftpClient, err := client.NewFTPClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer ftpClient.Disconnect()
    
    ctx := context.Background()
    if err := ftpClient.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    opts := &client.TransferOptions{
        Overwrite: true,
        Checksum:  true,
    }
    
    if err := ftpClient.Upload(ctx, "local.txt", "/remote/file.txt", opts); err != nil {
        log.Fatal(err)
    }
    
    log.Println("上传成功")
}
```

### 示例 2：带进度条的文件下载

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/f2xme/ftpx/pkg/client"
)

func main() {
    config := &client.Config{
        Protocol: client.ProtocolSFTP,
        Host:     "sftp.example.com",
        Port:     22,
        User:     "username",
        Auth: client.AuthConfig{
            Type:     client.AuthTypePassword,
            Password: "password",
        },
    }
    
    sftpClient, err := client.NewSFTPClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer sftpClient.Disconnect()
    
    ctx := context.Background()
    if err := sftpClient.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    opts := &client.TransferOptions{
        Resume:   true,
        Checksum: true,
        ProgressCallback: func(progress client.TransferProgress) {
            fmt.Printf("\r下载进度: %.2f%% [%d/%d bytes] %.2f MB/s 剩余: %s",
                progress.Percentage,
                progress.BytesTransferred,
                progress.TotalBytes,
                progress.Speed/1024/1024,
                progress.Remaining.Round(time.Second))
        },
    }
    
    if err := sftpClient.Download(ctx, "/remote/large.bin", "local.bin", opts); err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("\n下载完成")
}
```

### 示例 3：递归上传目录

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/f2xme/ftpx/pkg/client"
)

func main() {
    config := &client.Config{
        Protocol: client.ProtocolFTP,
        Host:     "ftp.example.com",
        Port:     21,
        User:     "username",
        Auth: client.AuthConfig{
            Type:     client.AuthTypePassword,
            Password: "password",
        },
        Options: client.Options{
            PassiveMode: true,
            Timeout:     60 * time.Second,
        },
    }
    
    ftpClient, err := client.NewFTPClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer ftpClient.Disconnect()
    
    ctx := context.Background()
    if err := ftpClient.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    opts := &client.TransferOptions{
        Recursive:     true,
        ParallelCount: 3,
        RateLimit:     5 * 1024 * 1024,  // 5 MB/s
        Checksum:      true,
    }
    
    if err := ftpClient.Upload(ctx, "./local-dir/", "/remote/dir/", opts); err != nil {
        log.Fatal(err)
    }
    
    log.Println("目录上传完成")
}
```

### 示例 4：批量文件列表

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/f2xme/ftpx/pkg/client"
)

func main() {
    config := &client.Config{
        Protocol: client.ProtocolSFTP,
        Host:     "sftp.example.com",
        Port:     22,
        User:     "username",
        Auth: client.AuthConfig{
            Type:    client.AuthTypeKey,
            KeyFile: "/home/user/.ssh/id_rsa",
        },
    }
    
    sftpClient, err := client.NewSFTPClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer sftpClient.Disconnect()
    
    ctx := context.Background()
    if err := sftpClient.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    files, err := sftpClient.List(ctx, "/remote/path")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("文件列表:")
    for _, file := range files {
        typeStr := "F"
        if file.IsDir {
            typeStr = "D"
        }
        fmt.Printf("[%s] %-40s %10d %s\n",
            typeStr,
            file.Name,
            file.Size,
            file.ModTime.Format("2006-01-02 15:04:05"))
    }
}
```

### 示例 5：目录同步

```go
package main

import (
    "context"
    "log"
    
    "github.com/f2xme/ftpx/pkg/client"
    "github.com/f2xme/ftpx/pkg/sync"
)

func main() {
    config := &client.Config{
        Protocol: client.ProtocolFTP,
        Host:     "ftp.example.com",
        Port:     21,
        User:     "username",
        Auth: client.AuthConfig{
            Type:     client.AuthTypePassword,
            Password: "password",
        },
    }
    
    ftpClient, err := client.NewFTPClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer ftpClient.Disconnect()
    
    ctx := context.Background()
    if err := ftpClient.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    syncOpts := &sync.Options{
        Bidirectional:     false,
        Delete:            true,
        ChecksumAlgorithm: "sha256",
        ExcludePatterns:   []string{"*.tmp", "*.log"},
        DryRun:            false,
    }
    
    syncer := sync.NewSyncer(ftpClient, syncOpts)
    result, err := syncer.Sync(ctx, "./local-dir/", "/remote/dir/")
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("同步完成: 上传=%d, 下载=%d, 删除=%d, 跳过=%d\n",
        result.Uploaded,
        result.Downloaded,
        result.Deleted,
        result.Skipped)
}
```

### 示例 6：连接池

```go
package main

import (
    "context"
    "log"
    "sync"
    
    "github.com/f2xme/ftpx/pkg/client"
    "github.com/f2xme/ftpx/pkg/pool"
)

func main() {
    config := &client.Config{
        Protocol: client.ProtocolFTP,
        Host:     "ftp.example.com",
        Port:     21,
        User:     "username",
        Auth: client.AuthConfig{
            Type:     client.AuthTypePassword,
            Password: "password",
        },
    }
    
    // 创建连接池
    p, err := pool.NewPool(config, 5, 10)  // 最小5，最大10个连接
    if err != nil {
        log.Fatal(err)
    }
    defer p.Close()
    
    // 并发上传多个文件
    var wg sync.WaitGroup
    files := []string{"file1.txt", "file2.txt", "file3.txt"}
    
    for _, file := range files {
        wg.Add(1)
        go func(filename string) {
            defer wg.Done()
            
            // 从池中获取客户端
            c, err := p.Get()
            if err != nil {
                log.Printf("获取连接失败: %v\n", err)
                return
            }
            defer p.Put(c)
            
            // 上传文件
            ctx := context.Background()
            opts := &client.TransferOptions{
                Overwrite: true,
            }
            
            if err := c.Upload(ctx, filename, "/remote/"+filename, opts); err != nil {
                log.Printf("上传 %s 失败: %v\n", filename, err)
                return
            }
            
            log.Printf("上传 %s 成功\n", filename)
        }(file)
    }
    
    wg.Wait()
    log.Println("所有文件上传完成")
}
```

### 示例 7：错误重试

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/f2xme/ftpx/pkg/client"
    "github.com/f2xme/ftpx/pkg/errors"
)

func uploadWithRetry(c client.Client, ctx context.Context, local, remote string, maxRetries int) error {
    opts := &client.TransferOptions{
        Resume:   true,
        Checksum: true,
    }
    
    for i := 0; i <= maxRetries; i++ {
        err := c.Upload(ctx, local, remote, opts)
        if err == nil {
            return nil
        }
        
        // 只重试网络错误
        if !errors.IsConnectionError(err) && !errors.IsTimeoutError(err) {
            return err
        }
        
        if i < maxRetries {
            waitTime := time.Duration(i+1) * 5 * time.Second
            log.Printf("上传失败，%s 后重试 (%d/%d): %v\n", waitTime, i+1, maxRetries, err)
            time.Sleep(waitTime)
            
            // 重新连接
            c.Disconnect()
            if err := c.Connect(ctx); err != nil {
                log.Printf("重新连接失败: %v\n", err)
                continue
            }
        }
    }
    
    return errors.New("达到最大重试次数")
}

func main() {
    config := &client.Config{
        Protocol: client.ProtocolFTP,
        Host:     "ftp.example.com",
        Port:     21,
        User:     "username",
        Auth: client.AuthConfig{
            Type:     client.AuthTypePassword,
            Password: "password",
        },
    }
    
    ftpClient, err := client.NewFTPClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer ftpClient.Disconnect()
    
    ctx := context.Background()
    if err := ftpClient.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    
    if err := uploadWithRetry(ftpClient, ctx, "large.bin", "/remote/large.bin", 3); err != nil {
        log.Fatal(err)
    }
    
    log.Println("上传成功")
}
```

## 最佳实践

### 1. 使用 Context 管理超时

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err := client.Upload(ctx, "local.txt", "/remote.txt", opts)
```

### 2. 适当设置缓冲区大小

```go
opts := &client.TransferOptions{
    BufferSize: 64 * 1024,  // 64KB，适合大多数场景
}
```

### 3. 使用连接池处理并发

```go
// 避免频繁创建和销毁连接
pool := pool.NewPool(config, 5, 10)
defer pool.Close()
```

### 4. 正确处理错误

```go
if err != nil {
    switch {
    case errors.IsConnectionError(err):
        // 重试连接
    case errors.IsNotFoundError(err):
        // 文件不存在，跳过
    default:
        // 其他错误，记录并返回
    }
}
```

### 5. 使用断点续传

```go
opts := &client.TransferOptions{
    Resume: true,  // 始终启用断点续传
}
```

### 6. 验证关键文件

```go
opts := &client.TransferOptions{
    Checksum:          true,
    ChecksumAlgorithm: "sha256",
}
```

## 性能调优

### 并发传输

```go
opts := &client.TransferOptions{
    ParallelCount: 5,  // 根据网络和服务器性能调整
}
```

### 速率限制

```go
opts := &client.TransferOptions{
    RateLimit: 5 * 1024 * 1024,  // 5 MB/s
}
```

### SFTP 压缩

```go
config.Options.Compression = true  // 网络慢时启用
```

## 线程安全

- `Client` 接口的实现**不是线程安全的**
- 并发场景请使用连接池或为每个 goroutine 创建独立的客户端
- 回调函数应该是线程安全的

## 资源管理

```go
// 总是使用 defer 确保资源释放
client, err := client.NewFTPClient(config)
if err != nil {
    return err
}
defer client.Disconnect()

// 或使用 context 控制生命周期
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
```
