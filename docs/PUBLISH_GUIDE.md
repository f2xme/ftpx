# FTPx 发布指南

## ✅ 已完成的工作

### 1. 项目重命名
- ✅ 项目名称：ftpcli → ftpx
- ✅ 目录名称：`/Users/bran/demo/ftpx`
- ✅ 二进制文件：`ftpx` (13.5MB)
- ✅ 配置目录：`~/.ftpx/`

### 2. 仓库配置
- ✅ Go Module：`github.com/f2xme/ftpx`
- ✅ 所有文件中的导入路径已更新
- ✅ Git 仓库已初始化
- ✅ 远程仓库已配置：`https://github.com/f2xme/ftpx.git`
- ✅ 初始提交已创建（commit: 627b474）

### 3. 代码完成度
- ✅ 所有核心功能实现
- ✅ FTP/FTPS/SFTP 支持
- ✅ 编译通过，功能测试通过
- ✅ 错误处理完善

### 4. 文档完成度
- ✅ README.md（项目主页）
- ✅ docs/USER_GUIDE.md（16,000+ 字用户指南）
- ✅ docs/QUICK_REFERENCE.md（快速参考）
- ✅ docs/API_REFERENCE.md（API 文档）
- ✅ 所有文档中的链接已更新

## 🚀 发布到 GitHub

### 步骤 1：在 GitHub 创建仓库

1. 访问 https://github.com/f2xme
2. 点击 "New repository"
3. 仓库名称：`ftpx`
4. 描述：`功能强大的跨平台 FTP/FTPS/SFTP 命令行客户端`
5. 选择 Public
6. **不要** 初始化 README、.gitignore 或 license（我们已有这些文件）
7. 点击 "Create repository"

### 步骤 2：推送代码

```bash
cd /Users/bran/demo/ftpx

# 如果需要配置 Git 用户信息
git config user.name "Your Name"
git config user.email "your.email@example.com"

# 推送到 GitHub
git push -u origin main

# 如果需要认证，可能需要：
# 1. 使用 GitHub Personal Access Token
# 2. 或配置 SSH 密钥
```

### 步骤 3：配置 GitHub 认证（如果需要）

#### 方式 A：使用 Personal Access Token

```bash
# 1. 在 GitHub 创建 Token
#    Settings → Developer settings → Personal access tokens → Generate new token
#    权限：repo (full control)

# 2. 使用 Token 推送
git push https://YOUR_TOKEN@github.com/f2xme/ftpx.git main
```

#### 方式 B：使用 SSH 密钥

```bash
# 1. 生成 SSH 密钥（如果没有）
ssh-keygen -t ed25519 -C "your.email@example.com"

# 2. 添加 SSH 密钥到 GitHub
#    Settings → SSH and GPG keys → New SSH key
cat ~/.ssh/id_ed25519.pub

# 3. 更改远程地址为 SSH
git remote set-url origin git@github.com:f2xme/ftpx.git

# 4. 推送
git push -u origin main
```

## 📋 发布后的任务

### 1. 完善 GitHub 仓库

- [ ] 添加仓库描述
- [ ] 设置 Topics（标签）：`ftp`, `sftp`, `ftps`, `cli`, `golang`, `file-transfer`
- [ ] 设置仓库 URL：添加项目网站或文档链接
- [ ] 配置 About 部分

### 2. 创建 Release

```bash
# 打标签
git tag -a v0.1.0 -m "Initial release: FTPx v0.1.0"
git push origin v0.1.0
```

在 GitHub 上创建 Release：
1. 进入仓库 → Releases → Create a new release
2. 选择标签：v0.1.0
3. Release 标题：`FTPx v0.1.0 - Initial Release`
4. 描述：参考 `RELEASE_NOTES_v0.2.0.md`
5. 上传编译好的二进制文件（可选）

### 3. 添加 Badges

在 README.md 顶部添加（已有占位符）：

```markdown
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/f2xme/ftpx)](https://github.com/f2xme/ftpx/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/f2xme/ftpx)](https://goreportcard.com/report/github.com/f2xme/ftpx)
```

### 4. 添加 LICENSE

创建 MIT License 文件：

```bash
cat > LICENSE << 'EOF'
MIT License

Copyright (c) 2026 f2xme

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF

git add LICENSE
git commit -m "Add MIT License"
git push
```

### 5. 配置 GitHub Actions（可选）

创建 `.github/workflows/build.yml`：

```yaml
name: Build and Test

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build
      run: go build -v ./...
    
    - name: Test
      run: go test -v ./...
```

## 📦 分发二进制文件

### 构建多平台版本

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o ftpx-linux-amd64 .

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o ftpx-linux-arm64 .

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o ftpx-darwin-amd64 .

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o ftpx-darwin-arm64 .

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o ftpx-windows-amd64.exe .
```

### 使用 GoReleaser（推荐）

创建 `.goreleaser.yml`：

```yaml
builds:
  - env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    binary: ftpx

archives:
  - format: tar.gz
    name_template: "ftpx_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

release:
  github:
    owner: f2xme
    name: ftpx
```

发布：
```bash
goreleaser release --clean
```

## 🎯 宣传和推广

### 1. 更新 README

确保 README 包含：
- ✅ 清晰的项目描述
- ✅ 功能特性列表
- ✅ 快速开始指南
- ✅ 使用示例
- ✅ 文档链接

### 2. 社区分享

考虑分享到：
- Reddit（r/golang, r/commandline）
- Hacker News
- Go Forum
- Twitter/X
- 中文社区（V2EX, 掘金）

### 3. 提交到工具列表

- Awesome Go (https://github.com/avelino/awesome-go)
- Go Projects (https://github.com/golang/go/wiki/Projects)

## 📊 监控和维护

### Issue 管理
- 及时回复用户问题
- 使用 labels 分类 issue
- 创建 issue templates

### Pull Request
- 设置 PR template
- Code review checklist
- 自动化测试

### 版本规划
- 使用语义化版本（Semantic Versioning）
- 维护 CHANGELOG.md
- 定期发布小版本

## ✅ 检查清单

发布前检查：

- [x] 代码编译通过
- [x] 核心功能测试通过
- [x] 文档完整
- [x] README 吸引人
- [x] 配置示例清晰
- [ ] GitHub 仓库已创建
- [ ] 代码已推送
- [ ] Release 已创建
- [ ] LICENSE 已添加
- [ ] CHANGELOG 已更新

## 🎉 完成！

项目现在已经准备好发布了。按照上述步骤推送到 GitHub，创建 Release，就可以让其他人使用你的工具了！

---

**项目地址**：https://github.com/f2xme/ftpx  
**最后更新**：2026-07-07  
**状态**：✅ 准备就绪
