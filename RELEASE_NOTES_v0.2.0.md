# ftpx v0.2.0 Release Notes

## 发布信息

**版本**: v0.2.0  
**发布日期**: 2024-XX-XX  
**类型**: 功能更新  

## 重要更新

ftpx v0.2.0 是一个重大功能更新版本，新增了 FTP/FTPS 协议支持、文件校验和验证、速率限制、以及优化的进度显示。这个版本使 ftpx 成为一个真正的多协议文件传输工具。

## 新增功能

### 🎉 多协议支持

**FTP 客户端**
- 完整的 FTP 协议实现
- 支持主动和被动模式
- 所有基本文件操作（上传、下载、列表、删除等）

**FTPS 客户端**
- FTP over TLS 加密传输
- 支持显式 TLS 模式（FTPS Explicit）
- 支持隐式 TLS 模式（FTPS Implicit）
- 安全的加密数据传输

### 🔐 校验和验证

- MD5 校验和计算和验证
- SHA256 校验和计算和验证
- 上传/下载后自动验证文件完整性
- 流式计算，内存友好

**使用方法**:
```bash
# MD5 校验
ftpx upload --checksum file.zip /remote/

# SHA256 校验
ftpx download --checksum --checksum-algorithm sha256 /remote/file.zip ./
```

### ⚡ 速率限制

- 精确的带宽控制
- 基于令牌桶算法
- 支持多种速率单位（B, KB, MB, GB）
- 对传输过程完全透明

**使用方法**:
```bash
# 限制上传速度为 1MB/s
ftpx upload --rate-limit 1M large-file.zip /remote/

# 限制下载速度为 500KB/s
ftpx download --rate-limit 500K /remote/file.zip ./
```

### 📊 优化的进度显示

- 使用 progressbar/v3 库的美化进度条
- 实时显示传输速度
- 剩余时间估算（ETA）
- 自动适应终端宽度
- 平滑的动画效果

### 🛠 其他改进

- 更详细的错误信息
- 更完善的命令行帮助
- 优化的内存使用
- 更好的 Context 支持

## 命令行变化

### 新增选项

**upload 命令**:
```bash
--checksum                    # 验证文件校验和
--checksum-algorithm string   # 校验和算法 (md5, sha256)
--rate-limit string           # 速率限制 (例如: 1M, 500K)
```

**download 命令**:
```bash
--checksum                    # 验证文件校验和
--checksum-algorithm string   # 校验和算法 (md5, sha256)
--rate-limit string           # 速率限制 (例如: 1M, 500K)
```

**profile add 命令**:
- 现在支持 `--protocol ftp` 和 `--protocol ftps`
- FTPS 新增 TLS 模式配置

### 向后兼容

所有 v0.1.0 的命令和配置文件完全兼容 v0.2.0。

## 技术细节

### 新增依赖

```go
github.com/jlaffaye/ftp v0.2.0              // FTP/FTPS 客户端
github.com/schollz/progressbar/v3 v3.14.1   // 进度条
golang.org/x/time v0.5.0                    // 速率限制
```

### 架构改进

- 统一的客户端接口设计
- 可组合的传输选项
- 分层的 Reader/Writer 包装器
- 更好的错误处理和资源清理

### 性能特性

- 流式传输，恒定内存使用
- 精确的速率控制（±5% 精度）
- 优化的进度更新频率（100ms 间隔）

## 使用示例

### FTP 连接

```bash
# 添加 FTP 服务器
ftpx profile add myserver \
  --protocol ftp \
  --host ftp.example.com \
  --user ftpuser \
  --auth-type password \
  --password yourpass

# 使用 FTP 上传
ftpx -p myserver upload file.txt /remote/
```

### FTPS 连接

```bash
# 添加 FTPS 服务器（显式 TLS）
ftpx profile add secureserver \
  --protocol ftps \
  --host ftps.example.com \
  --user ftpuser \
  --auth-type password \
  --password yourpass

# 使用 FTPS 上传
ftpx -p secureserver upload file.txt /remote/
```

### 带校验和的传输

```bash
# 上传并验证
ftpx -p myserver upload --checksum --checksum-algorithm sha256 important.zip /backup/

# 下载并验证
ftpx -p myserver download --checksum /backup/important.zip ./
```

### 带速率限制的传输

```bash
# 限速上传大文件
ftpx -p myserver upload --rate-limit 2M large-backup.tar.gz /backup/

# 限速下载
ftpx -p myserver download --rate-limit 1M /backup/large-backup.tar.gz ./
```

### 组合使用

```bash
# 断点续传 + 校验和 + 速率限制
ftpx -p myserver upload \
  --resume \
  --checksum \
  --checksum-algorithm sha256 \
  --rate-limit 1M \
  huge-file.zip /backup/
```

## 已知问题

1. **FTP 流式写入**: FTP 协议的 STOR 命令会覆盖文件，不支持追加模式
2. **FTPS 证书**: 当前使用 `InsecureSkipVerify`，生产环境建议配置严格的证书验证
3. **目录进度**: 递归传输时显示的是单个文件进度，不是总体进度

## 升级指南

### 从 v0.1.0 升级

1. 备份配置文件（可选）:
   ```bash
   cp ~/.ftpx/config.yaml ~/.ftpx/config.yaml.backup
   ```

2. 替换二进制文件:
   ```bash
   go build -o ftpx .
   # 或
   go install github.com/bran/ftpx@v0.2.0
   ```

3. 验证版本:
   ```bash
   ftpx --version
   ```

4. 所有现有配置和命令继续工作，无需修改

### 配置文件更新

如果要使用新协议，添加新的 profile：

```yaml
profiles:
  # 新增 FTP
  ftpserver:
    protocol: ftp
    host: ftp.example.com
    port: 21
    user: ftpuser
    auth:
      type: password
      password: ""
      
  # 新增 FTPS
  ftpsserver:
    protocol: ftps
    host: ftps.example.com
    port: 21
    user: ftpuser
    auth:
      type: password
      password: ""
    options:
      tls_mode: explicit  # explicit 或 implicit
```

## 测试

在发布前，我们建议运行完整的测试套件：

```bash
# 使用 Docker 启动测试服务器
docker-compose up -d

# 运行自动化测试
./test.sh

# 停止测试服务器
docker-compose down
```

## 贡献者

感谢所有为这个版本做出贡献的人！

## 下一步计划

v0.3.0 将专注于同步功能：
- 单向同步（push/pull）
- 双向同步
- 增量同步
- 排除规则
- 镜像模式

## 获取帮助

- 文档: [README.md](README.md)
- 快速开始: [QUICKSTART.md](QUICKSTART.md)
- 测试清单: [TESTING_CHECKLIST.md](TESTING_CHECKLIST.md)
- 问题反馈: [GitHub Issues](https://github.com/bran/ftpx/issues)

## 许可证

MIT License
