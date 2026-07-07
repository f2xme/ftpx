# ftpx Phase 2 测试验证清单

## 编译和基础功能测试

### ✅ 编译测试
- [x] 项目成功编译
- [x] 无编译错误
- [x] 无编译警告（忽略未使用参数提示）
- [x] 二进制文件生成成功

### ✅ 命令行测试
- [x] `ftpx --version` 显示版本信息
- [x] `ftpx --help` 显示帮助信息
- [x] `ftpx upload --help` 显示上传帮助
- [x] `ftpx download --help` 显示下载帮助
- [x] `ftpx ls --help` 显示列表帮助
- [x] `ftpx profile --help` 显示配置管理帮助

### ✅ 新功能选项确认
- [x] upload 命令包含 `--checksum` 选项
- [x] upload 命令包含 `--checksum-algorithm` 选项
- [x] upload 命令包含 `--rate-limit` 选项
- [x] download 命令包含 `--checksum` 选项
- [x] download 命令包含 `--checksum-algorithm` 选项
- [x] download 命令包含 `--rate-limit` 选项

## 需要实际服务器的功能测试

以下测试需要搭建实际的测试服务器：

### 📋 FTP 功能测试

**测试环境准备**:
```bash
# 使用 Docker 搭建 FTP 测试服务器
docker run -d \
  --name ftp-test \
  -p 21:21 \
  -p 21000-21010:21000-21010 \
  -e FTP_USER=testuser \
  -e FTP_PASS=testpass \
  fauria/vsftpd
```

**测试用例**:
- [ ] FTP 连接成功
- [ ] FTP 登录成功
- [ ] FTP 列出目录
- [ ] FTP 上传文件
- [ ] FTP 下载文件
- [ ] FTP 上传目录（递归）
- [ ] FTP 下载目录（递归）
- [ ] FTP 断点续传上传
- [ ] FTP 断点续传下载
- [ ] FTP 主动模式
- [ ] FTP 被动模式

### 📋 FTPS 功能测试

**测试环境准备**:
```bash
# 使用 Docker 搭建 FTPS 测试服务器
docker run -d \
  --name ftps-test \
  -p 21:21 \
  -p 21000-21010:21000-21010 \
  -e FTP_USER=testuser \
  -e FTP_PASS=testpass \
  -e FTPS_ENABLED=yes \
  fauria/vsftpd
```

**测试用例**:
- [ ] FTPS 显式 TLS 连接
- [ ] FTPS 隐式 TLS 连接
- [ ] FTPS 加密传输验证
- [ ] FTPS 证书验证
- [ ] FTPS 所有文件操作（与 FTP 相同）

### 📋 SFTP 功能测试

**测试环境准备**:
```bash
# 使用 Docker 搭建 SFTP 测试服务器
docker run -d \
  --name sftp-test \
  -p 2222:22 \
  -v /tmp/sftp-data:/home/testuser/upload \
  atmoz/sftp testuser:testpass:::upload
```

**测试用例**:
- [ ] SFTP 密码认证连接
- [ ] SFTP 密钥认证连接
- [ ] SFTP 列出目录
- [ ] SFTP 上传文件
- [ ] SFTP 下载文件
- [ ] SFTP 递归操作
- [ ] SFTP 断点续传

### 📋 校验和验证测试

**测试步骤**:
1. 创建测试文件并计算校验和
2. 使用 `--checksum` 上传文件
3. 手动验证远程文件校验和
4. 使用 `--checksum` 下载文件
5. 验证下载后文件校验和

**测试用例**:
- [ ] MD5 校验和正确计算
- [ ] SHA256 校验和正确计算
- [ ] 上传时校验和验证
- [ ] 下载时校验和验证
- [ ] 校验和不匹配时报错
- [ ] 大文件校验和（>100MB）

### 📋 速率限制测试

**测试步骤**:
1. 准备大文件（>10MB）
2. 使用不同速率限制上传/下载
3. 监控实际传输速率

**测试用例**:
- [ ] `--rate-limit 1M` 限制为 1MB/s
- [ ] `--rate-limit 500K` 限制为 500KB/s
- [ ] `--rate-limit 100K` 限制为 100KB/s
- [ ] 速率限制精度在 ±10% 范围内
- [ ] 无速率限制时全速传输
- [ ] 速率限制不影响小文件传输

### 📋 进度条显示测试

**测试用例**:
- [ ] 单文件上传显示进度条
- [ ] 单文件下载显示进度条
- [ ] 进度条显示百分比
- [ ] 进度条显示传输速度
- [ ] 进度条显示 ETA（如果支持）
- [ ] 目录传输显示合理的进度信息
- [ ] 小文件快速传输不闪烁
- [ ] 大文件平滑更新进度

### 📋 组合功能测试

测试多个功能同时使用：

- [ ] 断点续传 + 校验和验证
- [ ] 速率限制 + 进度条显示
- [ ] 递归上传 + 校验和验证
- [ ] 断点续传 + 速率限制
- [ ] 所有选项同时启用

### 📋 错误处理测试

- [ ] 连接失败正确报错
- [ ] 认证失败正确报错
- [ ] 文件不存在正确报错
- [ ] 权限不足正确报错
- [ ] 磁盘空间不足正确报错
- [ ] 网络中断自动重试/报错
- [ ] 校验和不匹配正确报错
- [ ] 无效的速率限制格式报错

### 📋 性能测试

- [ ] 小文件传输（<1MB）性能
- [ ] 中等文件传输（1-100MB）性能
- [ ] 大文件传输（>1GB）性能
- [ ] 内存使用稳定（不随文件大小增长）
- [ ] CPU 使用合理
- [ ] 多个文件传输性能

### 📋 兼容性测试

- [ ] 与常见 FTP 服务器兼容（vsftpd, ProFTPD）
- [ ] 与常见 SFTP 服务器兼容（OpenSSH）
- [ ] 不同操作系统（Linux, macOS, Windows）
- [ ] 不同终端环境（Terminal, iTerm2, Windows Terminal）

## 快速测试脚本

### 搭建所有测试服务器

```bash
#!/bin/bash

# 清理旧容器
docker rm -f ftp-test ftps-test sftp-test 2>/dev/null

# 启动 FTP 服务器
docker run -d \
  --name ftp-test \
  -p 21:21 \
  -p 21000-21010:21000-21010 \
  -e FTP_USER=testuser \
  -e FTP_PASS=testpass \
  fauria/vsftpd

# 启动 SFTP 服务器
docker run -d \
  --name sftp-test \
  -p 2222:22 \
  atmoz/sftp testuser:testpass:::upload

echo "测试服务器已启动！"
echo ""
echo "FTP:  localhost:21    (testuser/testpass)"
echo "SFTP: localhost:2222  (testuser/testpass)"
echo ""
echo "配置命令："
echo "ftpx profile add ftp-test --protocol ftp --host localhost --port 21 --user testuser --auth-type password --password testpass"
echo "ftpx profile add sftp-test --protocol sftp --host localhost --port 2222 --user testuser --auth-type password --password testpass"
```

### 基础功能测试脚本

```bash
#!/bin/bash

echo "=== ftpx 基础功能测试 ==="

# 创建测试文件
echo "创建测试文件..."
echo "Hello, ftpx!" > test-file.txt
mkdir -p test-dir
echo "File 1" > test-dir/file1.txt
echo "File 2" > test-dir/file2.txt

# 测试 SFTP
echo ""
echo "=== 测试 SFTP ==="
./ftpx -p sftp-test ls /upload
./ftpx -p sftp-test upload test-file.txt /upload/
./ftpx -p sftp-test upload -r test-dir /upload/
./ftpx -p sftp-test ls -lh /upload
./ftpx -p sftp-test download /upload/test-file.txt ./downloaded.txt

# 测试 FTP
echo ""
echo "=== 测试 FTP ==="
./ftpx -p ftp-test ls /
./ftpx -p ftp-test upload test-file.txt /
./ftpx -p ftp-test ls -lh /

# 测试校验和
echo ""
echo "=== 测试校验和 ==="
./ftpx -p sftp-test upload --checksum test-file.txt /upload/checksum-test.txt
./ftpx -p sftp-test download --checksum /upload/checksum-test.txt ./checksum-downloaded.txt

# 测试速率限制（需要大文件）
echo ""
echo "=== 测试速率限制 ==="
dd if=/dev/zero of=large-file.bin bs=1M count=10 2>/dev/null
./ftpx -p sftp-test upload --rate-limit 1M large-file.bin /upload/

echo ""
echo "=== 测试完成 ==="
echo "请检查上述命令输出是否正常"

# 清理
rm -f test-file.txt downloaded.txt checksum-downloaded.txt large-file.bin
rm -rf test-dir
```

## 已知限制

1. **FTP 流式写入**: FTP 协议的 STOR 命令会覆盖文件，不支持追加模式的流式写入
2. **证书验证**: FTPS 当前使用 `InsecureSkipVerify`，生产环境需要实现严格的证书验证
3. **目录进度**: 递归上传/下载目录时，进度条显示的是当前文件的进度，不是总体进度

## 测试报告模板

测试完成后，请填写以下报告：

```markdown
## ftpx v0.2.0 测试报告

**测试日期**: YYYY-MM-DD
**测试人**: 
**测试环境**: 
- OS: 
- Go 版本: 
- 服务器: 

### 测试结果

- FTP 功能: [ ] 通过 [ ] 失败
- FTPS 功能: [ ] 通过 [ ] 失败  
- SFTP 功能: [ ] 通过 [ ] 失败
- 校验和验证: [ ] 通过 [ ] 失败
- 速率限制: [ ] 通过 [ ] 失败
- 进度条显示: [ ] 通过 [ ] 失败

### 发现的问题

1. 
2. 
3. 

### 建议改进

1. 
2. 
3. 
```

## 下一步行动

1. ✅ 完成 Phase 2 开发
2. 📋 搭建测试环境
3. 📋 执行功能测试
4. 📋 修复发现的问题
5. 📋 发布 v0.2.0-beta
6. 📋 收集用户反馈
7. 📋 开始 Phase 3 开发
