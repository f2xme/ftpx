# FTPCli 使用指南

FTPCli 是一个功能强大的命令行 FTP/FTPS/SFTP 客户端工具，支持文件传输、目录管理、断点续传等功能。

## 目录

- [快速开始](#快速开始)
- [配置文件](#配置文件)
- [基本命令](#基本命令)
- [高级功能](#高级功能)
- [使用示例](#使用示例)
- [故障排查](#故障排查)

## 快速开始

### 安装

```bash
# 从源码编译
git clone <repository>
cd ftpx
go build -o ftpx .
```

### 初始化配置

首次运行时，会自动创建配置文件：

```bash
./ftpx config init
```

配置文件位置：`~/.ftpx/config.yaml`

### 添加服务器配置

```bash
# 添加 FTP 服务器
./ftpx config add myserver \
  --protocol ftp \
  --host ftp.example.com \
  --port 21 \
  --user username \
  --password yourpassword

# 添加 SFTP 服务器
./ftpx config add mysftp \
  --protocol sftp \
  --host sftp.example.com \
  --port 22 \
  --user username \
  --auth-type password \
  --password yourpassword

# 使用密钥认证
./ftpx config add mysftp-key \
  --protocol sftp \
  --host sftp.example.com \
  --user username \
  --auth-type key \
  --key-file ~/.ssh/id_rsa
```

## 配置文件

### 配置文件结构

```yaml
global:
  log_level: info                          # 日志级别：debug, info, warn, error
  log_file: ~/.ftpx/ftpx.log          # 日志文件路径
  default_profile: ""                      # 默认连接配置
  transfer:
    buffer_size: 32768                     # 传输缓冲区大小（字节）
    parallel_count: 3                      # 并发传输数
    timeout: 30s                           # 传输超时时间
    retry_count: 3                         # 重试次数
    rate_limit: 0                          # 全局速率限制（字节/秒，0 表示不限制）
  sync:
    checksum_algorithm: md5                # 校验和算法：md5, sha256
    ignore_patterns:                       # 同步时忽略的文件模式
      - .DS_Store
      - Thumbs.db
      - .git/

profiles:
  myserver:
    protocol: ftp                          # 协议：ftp, ftps, sftp
    host: ftp.example.com
    port: 21
    user: username
    auth:
      type: password                       # 认证类型：password, key
      password: yourpassword
    options:
      compression: false                   # 是否启用压缩（仅 SFTP）
      keep_alive: 30s                      # 保持连接活跃间隔
      passive_mode: true                   # 是否使用被动模式（FTP/FTPS）
      tls_skip_verify: false              # 是否跳过 TLS 证书验证（FTPS）
      tls_mode: explicit                   # TLS 模式：explicit, implicit（FTPS）
```

### 协议特定选项

#### FTP/FTPS
- `passive_mode`: 被动模式（推荐）
- `tls_mode`: TLS 模式
  - `explicit`: 显式 TLS（先明文连接，然后升级到 TLS）
  - `implicit`: 隐式 TLS（从一开始就使用 TLS）
- `tls_skip_verify`: 跳过证书验证（仅用于测试）

#### SFTP
- `compression`: 启用数据压缩
- `key_file`: SSH 密钥文件路径
- `key_password`: 密钥密码（如果密钥有密码保护）

## 基本命令

### 连接管理

```bash
# 列出所有配置
./ftpx config list

# 查看配置详情
./ftpx config show myserver

# 删除配置
./ftpx config remove myserver

# 测试连接
./ftpx -p myserver ping
```

### 文件和目录浏览

```bash
# 列出远程目录
./ftpx -p myserver ls /path/to/dir

# 详细列表（显示权限、大小、修改时间）
./ftpx -p myserver ls -l /path/to/dir

# 人类可读的文件大小
./ftpx -p myserver ls -l --human-readable /path/to/dir

# 查看当前工作目录
./ftpx -p myserver pwd

# 切换目录
./ftpx -p myserver cd /new/path
```

### 目录操作

```bash
# 创建目录
./ftpx -p myserver mkdir /path/to/newdir

# 递归创建目录
./ftpx -p myserver mkdir -r /path/to/deep/nested/dir

# 删除空目录
./ftpx -p myserver rmdir /path/to/dir

# 递归删除目录（包括内容）
./ftpx -p myserver rmdir -r /path/to/dir
```

### 文件操作

```bash
# 删除文件
./ftpx -p myserver rm /path/to/file.txt

# 重命名/移动文件
./ftpx -p myserver mv /old/path.txt /new/path.txt

# 查看文件详情
./ftpx -p myserver stat /path/to/file.txt

# 修改文件权限（仅 SFTP）
./ftpx -p myserver chmod 644 /path/to/file.txt
```

## 高级功能

### 文件上传

```bash
# 基本上传
./ftpx -p myserver upload local-file.txt /remote/path/file.txt

# 覆盖已存在文件
./ftpx -p myserver upload --overwrite local-file.txt /remote/path/file.txt

# 断点续传
./ftpx -p myserver upload --resume large-file.bin /remote/path/

# 速率限制（1MB/s）
./ftpx -p myserver upload --rate-limit 1M local-file.txt /remote/path/

# 上传后验证校验和
./ftpx -p myserver upload --checksum --checksum-algorithm sha256 file.txt /remote/

# 递归上传目录
./ftpx -p myserver upload -r local-dir/ /remote/dir/

# 并发上传（指定并发数）
./ftpx -p myserver upload --parallel 5 -r local-dir/ /remote/dir/

# 组合使用多个选项
./ftpx -p myserver upload \
  --rate-limit 2M \
  --checksum \
  --checksum-algorithm sha256 \
  --resume \
  --overwrite \
  large-file.bin /remote/path/
```

### 文件下载

```bash
# 基本下载
./ftpx -p myserver download /remote/file.txt local-file.txt

# 覆盖本地文件
./ftpx -p myserver download --overwrite /remote/file.txt local-file.txt

# 断点续传
./ftpx -p myserver download --resume /remote/large-file.bin local-file.bin

# 速率限制
./ftpx -p myserver download --rate-limit 500K /remote/file.txt ./

# 下载后验证校验和
./ftpx -p myserver download --checksum --checksum-algorithm md5 /remote/file.txt ./

# 递归下载目录
./ftpx -p myserver download -r /remote/dir/ local-dir/

# 组合使用
./ftpx -p myserver download \
  --rate-limit 10M \
  --checksum \
  --resume \
  -r /remote/backup/ ./local-backup/
```

### 速率限制

支持的速率格式：
- `500K` 或 `500KB`: 500 KB/s
- `1M` 或 `1MB`: 1 MB/s
- `10M`: 10 MB/s
- `512K`: 512 KB/s

```bash
# 限制上传速度为 1MB/s
./ftpx -p myserver upload --rate-limit 1M large-file.bin /remote/

# 限制下载速度为 500KB/s
./ftpx -p myserver download --rate-limit 500K /remote/file.bin ./
```

### 校验和验证

```bash
# 使用 MD5 校验和
./ftpx -p myserver upload --checksum file.txt /remote/

# 使用 SHA256 校验和
./ftpx -p myserver upload --checksum --checksum-algorithm sha256 file.txt /remote/

# 下载时验证
./ftpx -p myserver download --checksum --checksum-algorithm sha256 /remote/file.txt ./
```

### 断点续传

```bash
# 上传中断后继续
./ftpx -p myserver upload --resume partial-file.bin /remote/file.bin

# 下载中断后继续
./ftpx -p myserver download --resume /remote/large-file.bin local-file.bin
```

工作原理：
- 检查远程文件（上传）或本地文件（下载）的大小
- 从已传输的位置继续传输
- 仅传输剩余部分，节省时间和带宽

### 目录同步

```bash
# 同步本地目录到远程
./ftpx -p myserver sync local-dir/ /remote/dir/

# 双向同步
./ftpx -p myserver sync --bidirectional local-dir/ /remote/dir/

# 删除远程不存在于本地的文件
./ftpx -p myserver sync --delete local-dir/ /remote/dir/

# 使用 SHA256 校验和
./ftpx -p myserver sync --checksum-algorithm sha256 local-dir/ /remote/dir/

# 排除特定文件
./ftpx -p myserver sync --exclude "*.tmp" --exclude "*.log" local-dir/ /remote/dir/

# 仅同步指定文件类型
./ftpx -p myserver sync --include "*.pdf" local-dir/ /remote/dir/

# 干运行（预览更改）
./ftpx -p myserver sync --dry-run local-dir/ /remote/dir/
```

## 使用示例

### 场景 1：备份网站文件

```bash
# 下载整个网站目录
./ftpx -p webserver download -r /var/www/html ./backup/

# 使用校验和验证完整性
./ftpx -p webserver download \
  -r \
  --checksum \
  --checksum-algorithm sha256 \
  /var/www/html ./backup/
```

### 场景 2：上传大文件

```bash
# 上传大文件，限速并支持断点续传
./ftpx -p myserver upload \
  --rate-limit 2M \
  --resume \
  --checksum \
  large-backup.tar.gz /backups/

# 如果上传中断，再次运行相同命令继续
./ftpx -p myserver upload \
  --rate-limit 2M \
  --resume \
  --checksum \
  large-backup.tar.gz /backups/
```

### 场景 3：自动化部署

```bash
#!/bin/bash
# deploy.sh - 自动部署脚本

# 1. 构建项目
npm run build

# 2. 上传构建产物
./ftpx -p production upload \
  -r \
  --overwrite \
  --checksum \
  ./dist/ /var/www/production/

# 3. 验证部署
./ftpx -p production ls -l /var/www/production/
```

### 场景 4：目录同步

```bash
# 保持本地和远程同步
./ftpx -p myserver sync \
  --bidirectional \
  --checksum-algorithm sha256 \
  ./documents/ /remote/documents/

# 单向同步（本地 → 远程），删除远程多余文件
./ftpx -p myserver sync \
  --delete \
  ./source/ /remote/destination/
```

### 场景 5：批量文件传输

```bash
# 递归上传多个目录，限速避免占满带宽
./ftpx -p myserver upload \
  -r \
  --rate-limit 5M \
  --parallel 3 \
  ./data/ /remote/data/

# 递归下载，验证每个文件
./ftpx -p myserver download \
  -r \
  --checksum \
  --checksum-algorithm md5 \
  /remote/archive/ ./local-archive/
```

## 故障排查

### 连接问题

**无法连接到服务器**
```bash
# 1. 测试连接
./ftpx -p myserver ping

# 2. 检查配置
./ftpx config show myserver

# 3. 使用详细日志
./ftpx -v -p myserver ls /
```

**被动模式连接失败（FTP）**
```yaml
# 在配置中禁用被动模式
options:
  passive_mode: false
```

**FTPS 证书验证失败**
```yaml
# 临时跳过证书验证（仅用于测试）
options:
  tls_skip_verify: true
```

### 传输问题

**上传/下载速度慢**
```bash
# 1. 检查是否有速率限制
./ftpx config show myserver

# 2. 增加并发数
./ftpx -p myserver upload --parallel 5 -r ./dir/ /remote/

# 3. 启用压缩（仅 SFTP）
# 在配置文件中设置 compression: true
```

**断点续传不工作**
- 确保使用 `--resume` 标志
- 检查远程文件权限
- 确认服务器支持断点续传

**校验和不匹配**
- 文件在传输过程中被修改
- 网络传输错误
- 重新上传/下载文件

### 权限问题

**无法创建目录**
```bash
# 检查远程目录权限
./ftpx -p myserver ls -l /parent/directory/

# 确保有写权限
# FTP/FTPS: 确保服务器配置允许写入
# SFTP: 检查 SSH 用户权限
```

**递归上传失败**
- 确保远程目录可写
- 检查磁盘空间
- 验证文件名是否包含特殊字符

### 性能优化

**大文件传输优化**
```bash
# 1. 使用断点续传
--resume

# 2. 调整缓冲区大小（在配置文件中）
transfer:
  buffer_size: 65536  # 64KB

# 3. 合理设置速率限制
--rate-limit 10M
```

**批量文件传输优化**
```bash
# 1. 增加并发数
--parallel 5

# 2. 对于 SFTP，启用压缩
compression: true

# 3. 调整超时时间
transfer:
  timeout: 60s
```

## 环境变量

```bash
# 配置文件路径
export FTPCLI_CONFIG=~/.ftpx/config.yaml

# 默认配置文件
export FTPCLI_PROFILE=myserver

# 日志级别
export FTPCLI_LOG_LEVEL=debug
```

## 命令行标志

### 全局标志
- `--config <path>`: 指定配置文件路径
- `-p, --profile <name>`: 使用指定的连接配置
- `-v, --verbose`: 详细输出

### 传输标志
- `-r, --recursive`: 递归处理目录
- `--overwrite`: 覆盖已存在文件
- `--resume`: 断点续传
- `--rate-limit <speed>`: 速率限制
- `--parallel <count>`: 并发传输数
- `--checksum`: 验证校验和
- `--checksum-algorithm <alg>`: 校验和算法（md5, sha256）

### 同步标志
- `--bidirectional`: 双向同步
- `--delete`: 删除目标中多余的文件
- `--dry-run`: 预览更改，不实际执行
- `--exclude <pattern>`: 排除文件模式
- `--include <pattern>`: 仅包含文件模式

## 最佳实践

1. **使用配置文件管理多个服务器**
   - 避免在命令行中暴露密码
   - 使用有意义的配置名称

2. **传输大文件时使用断点续传**
   - 防止网络中断导致重新传输
   - 节省时间和带宽

3. **验证重要文件的完整性**
   - 使用 `--checksum` 验证传输正确性
   - 对于关键数据使用 SHA256

4. **合理设置速率限制**
   - 避免占满带宽影响其他服务
   - 在生产环境使用适当的限制

5. **使用同步而不是手动上传/下载**
   - 自动处理增量更新
   - 保持目录一致性

6. **定期备份配置文件**
   ```bash
   cp ~/.ftpx/config.yaml ~/.ftpx/config.yaml.backup
   ```

7. **使用密钥认证而不是密码（SFTP）**
   - 更安全
   - 支持自动化脚本

## 技术支持

- 问题反馈：[GitHub Issues](https://github.com/your-repo/ftpx/issues)
- 文档：[在线文档](https://docs.example.com)
- 示例：[examples/](../examples/)
