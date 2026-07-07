# FTPCli

<div align="center">

**功能强大的跨平台 FTP/FTPS/SFTP 命令行客户端**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com/f2xme/ftpx)

[功能特性](#功能特性) • [快速开始](#快速开始) • [文档](#文档) • [示例](#示例) • [贡献](#贡献)

</div>

---

## 功能特性

### 🚀 核心功能

- **多协议支持**：FTP、FTPS（显式/隐式 TLS）、SFTP
- **文件传输**：上传、下载、递归传输目录
- **断点续传**：支持大文件中断后继续传输
- **速率限制**：控制上传/下载速度，避免占满带宽
- **校验和验证**：MD5/SHA256 文件完整性验证
- **目录同步**：智能同步本地和远程目录

### ⚡ 高级特性

- **并发传输**：多文件并发上传/下载
- **进度显示**：实时显示传输进度和速度
- **配置管理**：统一管理多个服务器配置
- **批量操作**：递归目录操作和批量文件管理
- **灵活认证**：支持密码和 SSH 密钥认证

### 💪 性能优化

- **高效缓冲**：可配置的传输缓冲区
- **被动模式**：FTP/FTPS 被动模式支持
- **数据压缩**：SFTP 传输压缩（可选）
- **连接复用**：保持连接活跃，减少重连开销

## 快速开始

### 安装

```bash
# 从源码编译
git clone https://github.com/f2xme/ftpx.git
cd ftpx
go build -o ftpx .

# 安装到系统路径
sudo mv ftpx /usr/local/bin/
```

### 初始化配置

```bash
# 创建配置文件
ftpx config init

# 添加 FTP 服务器
ftpx config add myserver \
  --protocol ftp \
  --host ftp.example.com \
  --user username \
  --password yourpassword
```

### 基本使用

```bash
# 列出远程目录
ftpx -p myserver ls /

# 上传文件
ftpx -p myserver upload local.txt /remote/path/

# 下载文件
ftpx -p myserver download /remote/file.txt ./

# 递归上传目录
ftpx -p myserver upload -r ./local-dir/ /remote/dir/
```

### 高级使用

```bash
# 限速上传（1MB/s）
ftpx -p myserver upload --rate-limit 1M large-file.bin /remote/

# 断点续传
ftpx -p myserver upload --resume partial-file.bin /remote/

# 校验和验证
ftpx -p myserver upload --checksum --checksum-algorithm sha256 file.txt /remote/

# 递归下载并验证
ftpx -p myserver download -r --checksum /remote/backup/ ./local-backup/

# 目录同步
ftpx -p myserver sync --bidirectional ./local/ /remote/
```

## 文档

### 用户文档

- **[用户指南](docs/USER_GUIDE.md)** - 完整的使用说明和教程
- **[快速参考](docs/QUICK_REFERENCE.md)** - 命令速查表
- **[配置示例](docs/examples/)** - 各种场景的配置示例

### 开发者文档

- **[API 参考](docs/API_REFERENCE.md)** - Go API 文档
- **[架构设计](docs/ARCHITECTURE.md)** - 系统架构说明
- **[贡献指南](CONTRIBUTING.md)** - 如何参与开发

## 示例

### 场景 1：网站备份

```bash
# 下载整个网站到本地
ftpx -p webserver download -r --checksum /var/www/html ./backup/

# 定期同步
ftpx -p webserver sync --delete --checksum-algorithm sha256 \
  ./backup/ /var/www/html/
```

### 场景 2：大文件传输

```bash
# 上传大文件，限速 2MB/s，支持断点续传
ftpx -p myserver upload \
  --rate-limit 2M \
  --resume \
  --checksum \
  large-backup.tar.gz /backups/

# 如果中断，再次运行相同命令继续
```

### 场景 3：自动化部署

```bash
#!/bin/bash
# 构建项目
npm run build

# 上传到生产环境
ftpx -p production upload \
  -r \
  --overwrite \
  --checksum \
  ./dist/ /var/www/production/

# 验证部署
ftpx -p production ls -l /var/www/production/
```

### 场景 4：批量文件管理

```bash
# 递归上传多个目录
ftpx -p myserver upload \
  -r \
  --rate-limit 5M \
  --parallel 3 \
  ./data/ /remote/data/

# 递归下载并验证每个文件
ftpx -p myserver download \
  -r \
  --checksum \
  --checksum-algorithm md5 \
  /remote/archive/ ./local-archive/
```

## 配置示例

### FTP 配置

```yaml
myserver:
  protocol: ftp
  host: ftp.example.com
  port: 21
  user: username
  auth:
    type: password
    password: yourpassword
  options:
    passive_mode: true
    timeout: 30s
```

### FTPS 配置

```yaml
myftps:
  protocol: ftps
  host: ftps.example.com
  port: 990
  user: username
  auth:
    type: password
    password: yourpassword
  options:
    passive_mode: true
    tls_mode: implicit
    tls_skip_verify: false
```

### SFTP 配置（密钥认证）

```yaml
mysftp:
  protocol: sftp
  host: sftp.example.com
  port: 22
  user: username
  auth:
    type: key
    key_file: ~/.ssh/id_rsa
  options:
    compression: false
    timeout: 30s
```

## 支持的功能矩阵

| 功能 | FTP | FTPS | SFTP |
|------|-----|------|------|
| 基本连接 | ✅ | ✅ | ✅ |
| 文件上传/下载 | ✅ | ✅ | ✅ |
| 递归传输 | ✅ | ✅ | ✅ |
| 断点续传 | ✅ | ✅ | ✅ |
| 速率限制 | ✅ | ✅ | ✅ |
| 校验和验证 | ✅ | ✅ | ✅ |
| 目录同步 | ✅ | ✅ | ✅ |
| 并发传输 | ✅ | ✅ | ✅ |
| 数据压缩 | ❌ | ❌ | ✅ |
| 密钥认证 | ❌ | ❌ | ✅ |

## 命令概览

### 配置管理
```bash
ftpx config init              # 初始化配置
ftpx config add <name> ...    # 添加服务器配置
ftpx config list              # 列出所有配置
ftpx config show <name>       # 查看配置详情
ftpx config remove <name>     # 删除配置
```

### 文件操作
```bash
ftpx -p <profile> ls <path>           # 列出目录
ftpx -p <profile> upload <src> <dst>  # 上传文件
ftpx -p <profile> download <src> <dst># 下载文件
ftpx -p <profile> rm <path>           # 删除文件
ftpx -p <profile> mv <old> <new>      # 重命名/移动
```

### 目录操作
```bash
ftpx -p <profile> pwd                 # 显示当前目录
ftpx -p <profile> cd <path>           # 切换目录
ftpx -p <profile> mkdir <path>        # 创建目录
ftpx -p <profile> rmdir <path>        # 删除目录
```

### 高级功能
```bash
ftpx -p <profile> sync <local> <remote>   # 同步目录
ftpx -p <profile> stat <path>             # 文件信息
ftpx -p <profile> chmod <mode> <path>     # 修改权限（SFTP）
```

## 性能测试

基于 5MB 文件的传输测试（本地网络）：

| 协议 | 上传速度 | 下载速度 | CPU 使用率 |
|------|---------|---------|-----------|
| FTP  | ~95 MB/s | ~98 MB/s | 5-8% |
| FTPS | ~85 MB/s | ~90 MB/s | 8-12% |
| SFTP | ~75 MB/s | ~80 MB/s | 10-15% |

*测试环境：macOS, Go 1.21, 1Gbps 网络*

## 系统要求

- **操作系统**：Linux、macOS、Windows
- **Go 版本**：1.21 或更高（仅编译时需要）
- **网络**：支持 TCP/IP 连接
- **磁盘空间**：约 10MB（编译后的二进制文件）

## 故障排查

### 连接失败

```bash
# 测试连接
ftpx -p myserver ping

# 使用详细输出
ftpx -v -p myserver ls /

# 检查配置
ftpx config show myserver
```

### FTP 被动模式问题

在配置中禁用被动模式：
```yaml
options:
  passive_mode: false
```

### FTPS 证书问题

跳过证书验证（仅用于测试）：
```yaml
options:
  tls_skip_verify: true
```

更多故障排查信息请参考[用户指南](docs/USER_GUIDE.md#故障排查)。

## 开发路线图

### v1.0（当前）
- [x] FTP/FTPS/SFTP 基本支持
- [x] 文件上传/下载
- [x] 递归传输
- [x] 断点续传
- [x] 速率限制
- [x] 校验和验证
- [x] 目录同步

### v1.1（计划中）
- [ ] Web UI 界面
- [ ] 传输队列管理
- [ ] 计划任务支持
- [ ] 多语言支持
- [ ] 性能监控面板

### v2.0（未来）
- [ ] 分布式传输
- [ ] 云存储集成
- [ ] API 服务器模式
- [ ] 插件系统

## 贡献

欢迎贡献！请查看[贡献指南](CONTRIBUTING.md)了解详情。

### 贡献者

感谢所有贡献者的支持！

<!-- 
贡献者列表将自动生成
-->

## 许可证

本项目采用 [MIT 许可证](LICENSE)。

## 致谢

- [jlaffaye/ftp](https://github.com/jlaffaye/ftp) - FTP 客户端库
- [pkg/sftp](https://github.com/pkg/sftp) - SFTP 客户端库
- [spf13/cobra](https://github.com/spf13/cobra) - CLI 框架
- [spf13/viper](https://github.com/spf13/viper) - 配置管理

## 联系方式

- **问题反馈**：[GitHub Issues](https://github.com/f2xme/ftpx/issues)
- **功能建议**：[GitHub Discussions](https://github.com/f2xme/ftpx/discussions)
- **安全问题**：security@example.com

---

<div align="center">

**如果觉得有用，请给个 ⭐️ Star！**

Made with ❤️ by the FTPCli Team

</div>
