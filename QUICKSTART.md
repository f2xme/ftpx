# ftpx 快速使用指南

## 编译

```bash
cd ftpx
go build -o ftpx .
```

## 基本使用流程

### 1. 添加 SFTP 服务器配置

**使用密钥认证（推荐）：**
```bash
./ftpx profile add myserver \
  --protocol sftp \
  --host example.com \
  --port 22 \
  --user admin \
  --auth-type key \
  --key-file ~/.ssh/id_rsa
```

**使用密码认证：**
```bash
./ftpx profile add myserver \
  --protocol sftp \
  --host example.com \
  --port 22 \
  --user admin \
  --auth-type password \
  --password yourpassword
```

### 2. 查看所有配置

```bash
./ftpx profile list
```

### 3. 查看配置详情

```bash
./ftpx profile show myserver
```

### 4. 列出远程目录

```bash
# 简单列表
./ftpx -p myserver ls /path

# 详细信息
./ftpx -p myserver ls -l /path

# 详细信息 + 人类可读大小
./ftpx -p myserver ls -lh /path
```

### 5. 上传文件

```bash
# 上传单个文件
./ftpx -p myserver upload file.txt /remote/path/

# 上传目录（递归）
./ftpx -p myserver upload -r ./local-dir /remote/dir

# 断点续传
./ftpx -p myserver upload --resume large-file.zip /remote/

# 覆盖已存在文件
./ftpx -p myserver upload --overwrite file.txt /remote/
```

### 6. 下载文件

```bash
# 下载单个文件
./ftpx -p myserver download /remote/file.txt ./

# 下载目录（递归）
./ftpx -p myserver download -r /remote/dir ./local-dir

# 断点续传
./ftpx -p myserver download --resume /remote/large-file.zip ./

# 覆盖已存在文件
./ftpx -p myserver download --overwrite /remote/file.txt ./
```

### 7. 删除配置

```bash
./ftpx profile remove myserver
```

## 配置文件

配置文件会自动创建在 `~/.ftpx/config.yaml`。

你也可以手动创建配置文件，参考 `config.example.yaml`。

## 常见问题

### Q: 如何使用带密码的 SSH 密钥？

```bash
./ftpx profile add myserver \
  --protocol sftp \
  --host example.com \
  --user admin \
  --auth-type key \
  --key-file ~/.ssh/id_rsa \
  --passphrase your-key-password
```

### Q: 如何在不保存密码的情况下使用密码认证？

将密码留空，运行时会提示输入：

```bash
./ftpx profile add myserver \
  --protocol sftp \
  --host example.com \
  --user admin \
  --auth-type password
```

或者在配置文件中将 password 字段留空。

### Q: 如何启用详细输出？

在任何命令中添加 `-v` 或 `--verbose` 标志：

```bash
./ftpx -v -p myserver ls /path
```

### Q: 如何更改默认配置？

编辑 `~/.ftpx/config.yaml`，设置 `global.default_profile` 字段：

```yaml
global:
  default_profile: "myserver"
```

这样就不需要每次都使用 `-p` 参数了。

## 实际使用示例

### 备份网站到本地

```bash
# 下载整个网站目录
./ftpx -p prod download -r /var/www/html ./website-backup
```

### 部署静态网站

```bash
# 上传构建后的文件
./ftpx -p prod upload -r ./dist /var/www/html
```

### 下载日志文件

```bash
# 下载日志目录
./ftpx -p prod download -r /var/log/nginx ./logs
```

### 上传大文件（支持断点续传）

```bash
# 如果网络中断，可以重新运行相同命令继续传输
./ftpx -p prod upload --resume backup-2024.tar.gz /backups/
```

## 当前功能状态

✅ **已实现：**
- SFTP 连接（密码/密钥认证）
- 文件上传/下载
- 目录递归操作
- 断点续传
- 进度显示
- 配置管理
- 列出目录

🚧 **开发中：**
- FTP/FTPS 支持
- 目录同步
- 并发传输优化
- 交互式 shell 模式

📋 **计划中：**
- 速率限制
- 校验和验证
- 更丰富的进度条
- 批量操作
- 历史记录

## 技术支持

遇到问题或有建议？欢迎：
- 提交 Issue
- 发起 Pull Request
- 查看详细文档: [README.md](README.md)
