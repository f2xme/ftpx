# FTPCli 快速参考

## 命令速查表

### 配置管理

```bash
# 初始化配置
ftpx config init

# 添加 FTP 服务器
ftpx config add myserver --protocol ftp --host ftp.example.com --user username --password pass

# 添加 SFTP 服务器（密码认证）
ftpx config add mysftp --protocol sftp --host sftp.example.com --user username --password pass

# 添加 SFTP 服务器（密钥认证）
ftpx config add mysftp --protocol sftp --host sftp.example.com --user username --auth-type key --key-file ~/.ssh/id_rsa

# 列出所有配置
ftpx config list

# 查看配置详情
ftpx config show myserver

# 删除配置
ftpx config remove myserver

# 测试连接
ftpx -p myserver ping
```

### 浏览和导航

```bash
# 列出目录
ftpx -p myserver ls /path/to/dir

# 详细列表
ftpx -p myserver ls -l /path/to/dir

# 人类可读的大小
ftpx -p myserver ls -l --human-readable /path/to/dir

# 当前工作目录
ftpx -p myserver pwd

# 切换目录
ftpx -p myserver cd /new/path
```

### 目录操作

```bash
# 创建目录
ftpx -p myserver mkdir /path/to/dir

# 递归创建
ftpx -p myserver mkdir -r /path/to/deep/dir

# 删除空目录
ftpx -p myserver rmdir /path/to/dir

# 递归删除
ftpx -p myserver rmdir -r /path/to/dir
```

### 文件操作

```bash
# 删除文件
ftpx -p myserver rm /path/to/file.txt

# 重命名/移动
ftpx -p myserver mv /old/path.txt /new/path.txt

# 查看文件信息
ftpx -p myserver stat /path/to/file.txt

# 修改权限（SFTP）
ftpx -p myserver chmod 644 /path/to/file.txt
```

### 上传

```bash
# 基本上传
ftpx -p myserver upload local.txt /remote/path/

# 递归上传目录
ftpx -p myserver upload -r local-dir/ /remote/dir/

# 覆盖已存在文件
ftpx -p myserver upload --overwrite file.txt /remote/

# 断点续传
ftpx -p myserver upload --resume large-file.bin /remote/

# 速率限制（1MB/s）
ftpx -p myserver upload --rate-limit 1M file.txt /remote/

# 验证校验和
ftpx -p myserver upload --checksum file.txt /remote/

# 使用 SHA256
ftpx -p myserver upload --checksum --checksum-algorithm sha256 file.txt /remote/

# 并发上传
ftpx -p myserver upload --parallel 5 -r dir/ /remote/

# 组合使用
ftpx -p myserver upload -r --rate-limit 2M --checksum --resume local/ /remote/
```

### 下载

```bash
# 基本下载
ftpx -p myserver download /remote/file.txt local.txt

# 递归下载目录
ftpx -p myserver download -r /remote/dir/ local-dir/

# 覆盖本地文件
ftpx -p myserver download --overwrite /remote/file.txt ./

# 断点续传
ftpx -p myserver download --resume /remote/large-file.bin local.bin

# 速率限制
ftpx -p myserver download --rate-limit 500K /remote/file.txt ./

# 验证校验和
ftpx -p myserver download --checksum /remote/file.txt ./

# 组合使用
ftpx -p myserver download -r --rate-limit 10M --checksum /remote/ ./local/
```

### 同步

```bash
# 单向同步（本地 → 远程）
ftpx -p myserver sync local-dir/ /remote/dir/

# 双向同步
ftpx -p myserver sync --bidirectional local-dir/ /remote/dir/

# 删除目标中多余文件
ftpx -p myserver sync --delete local-dir/ /remote/dir/

# 使用 SHA256 校验
ftpx -p myserver sync --checksum-algorithm sha256 local-dir/ /remote/dir/

# 排除文件
ftpx -p myserver sync --exclude "*.tmp" --exclude "*.log" local/ /remote/

# 仅包含特定文件
ftpx -p myserver sync --include "*.pdf" local/ /remote/

# 干运行（预览）
ftpx -p myserver sync --dry-run local/ /remote/
```

## 常用选项

| 选项 | 说明 | 示例 |
|------|------|------|
| `-p, --profile` | 使用指定配置 | `-p myserver` |
| `-v, --verbose` | 详细输出 | `-v` |
| `-r, --recursive` | 递归处理目录 | `-r` |
| `--overwrite` | 覆盖已存在文件 | `--overwrite` |
| `--resume` | 断点续传 | `--resume` |
| `--rate-limit` | 速率限制 | `--rate-limit 1M` |
| `--parallel` | 并发数 | `--parallel 5` |
| `--checksum` | 验证校验和 | `--checksum` |
| `--checksum-algorithm` | 校验算法 | `--checksum-algorithm sha256` |
| `--dry-run` | 预览不执行 | `--dry-run` |

## 速率限制格式

| 格式 | 说明 |
|------|------|
| `500K` | 500 KB/s |
| `1M` | 1 MB/s |
| `10M` | 10 MB/s |
| `512K` | 512 KB/s |

## 校验和算法

| 算法 | 说明 | 使用场景 |
|------|------|----------|
| `md5` | MD5（默认） | 快速验证 |
| `sha256` | SHA-256 | 安全性要求高 |

## 协议配置

### FTP

```yaml
myserver:
  protocol: ftp
  host: ftp.example.com
  port: 21
  user: username
  auth:
    type: password
    password: yourpass
  options:
    passive_mode: true
```

### FTPS

```yaml
myftps:
  protocol: ftps
  host: ftps.example.com
  port: 990
  user: username
  auth:
    type: password
    password: yourpass
  options:
    passive_mode: true
    tls_mode: implicit          # explicit 或 implicit
    tls_skip_verify: false
```

### SFTP（密码）

```yaml
mysftp:
  protocol: sftp
  host: sftp.example.com
  port: 22
  user: username
  auth:
    type: password
    password: yourpass
  options:
    compression: false
```

### SFTP（密钥）

```yaml
mysftp:
  protocol: sftp
  host: sftp.example.com
  port: 22
  user: username
  auth:
    type: key
    key_file: ~/.ssh/id_rsa
    key_password: ""             # 如果密钥有密码
  options:
    compression: false
```

## 常见任务

### 备份网站

```bash
# 下载整个网站
ftpx -p webserver download -r --checksum /var/www/html ./backup/

# 定期同步
ftpx -p webserver sync --delete --checksum-algorithm sha256 ./backup/ /var/www/html/
```

### 上传大文件

```bash
# 限速、断点续传、校验
ftpx -p myserver upload \
  --rate-limit 2M \
  --resume \
  --checksum \
  --checksum-algorithm sha256 \
  large-backup.tar.gz /backups/
```

### 批量传输

```bash
# 递归上传，限速，并发
ftpx -p myserver upload \
  -r \
  --rate-limit 5M \
  --parallel 5 \
  --checksum \
  ./data/ /remote/data/
```

### 自动部署

```bash
#!/bin/bash
# 构建
npm run build

# 上传
ftpx -p production upload \
  -r \
  --overwrite \
  --checksum \
  ./dist/ /var/www/production/

# 验证
ftpx -p production ls -l /var/www/production/
```

## 故障排查

### 连接问题

```bash
# 测试连接
ftpx -p myserver ping

# 详细日志
ftpx -v -p myserver ls /

# 检查配置
ftpx config show myserver
```

### FTP 被动模式问题

```yaml
# 配置文件中禁用被动模式
options:
  passive_mode: false
```

### FTPS 证书问题

```yaml
# 跳过证书验证（仅测试）
options:
  tls_skip_verify: true
```

### 权限问题

```bash
# 检查权限
ftpx -p myserver ls -l /parent/directory/

# SFTP 修改权限
ftpx -p myserver chmod 755 /path/to/dir
```

## 性能优化

### 大文件传输

- 使用 `--resume` 断点续传
- 调整缓冲区：`buffer_size: 65536`
- 合理限速：`--rate-limit 10M`

### 批量文件

- 增加并发：`--parallel 5`
- SFTP 压缩：`compression: true`
- 延长超时：`timeout: 60s`

### 网络优化

- FTP：使用被动模式
- SFTP：启用压缩
- 避免过高并发数

## 环境变量

```bash
# 配置文件路径
export FTPCLI_CONFIG=~/.ftpx/config.yaml

# 默认配置
export FTPCLI_PROFILE=myserver

# 日志级别
export FTPCLI_LOG_LEVEL=debug
```

## 退出代码

| 代码 | 说明 |
|------|------|
| 0 | 成功 |
| 1 | 一般错误 |
| 2 | 连接失败 |
| 3 | 认证失败 |
| 4 | 文件不存在 |
| 5 | 权限拒绝 |

## 日志级别

| 级别 | 说明 |
|------|------|
| `debug` | 调试信息（最详细） |
| `info` | 一般信息 |
| `warn` | 警告信息 |
| `error` | 错误信息（最少） |

## 配置文件位置

- 默认：`~/.ftpx/config.yaml`
- 自定义：`--config /path/to/config.yaml`
- 环境变量：`$FTPCLI_CONFIG`

## 日志文件位置

- 默认：`~/.ftpx/ftpx.log`
- 配置：`log_file: /path/to/logfile.log`

## 帮助命令

```bash
# 查看所有命令
ftpx --help

# 查看特定命令帮助
ftpx upload --help
ftpx download --help
ftpx sync --help
ftpx config --help
```
