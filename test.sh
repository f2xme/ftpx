#!/bin/bash

# ftpcli 自动化测试脚本

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 计数器
TESTS_PASSED=0
TESTS_FAILED=0

# 打印函数
print_header() {
    echo ""
    echo "=================================================="
    echo "$1"
    echo "=================================================="
}

print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

# 检查 ftpcli 是否存在
if [ ! -f "./ftpcli" ]; then
    echo -e "${RED}错误: ./ftpcli 不存在，请先编译项目${NC}"
    exit 1
fi

# 检查 Docker 是否运行
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}错误: Docker 未运行，请启动 Docker${NC}"
    exit 1
fi

print_header "步骤 1: 启动测试服务器"

echo "启动 Docker 容器..."
docker-compose up -d

echo "等待服务器启动 (10秒)..."
sleep 10

print_pass "测试服务器已启动"

print_header "步骤 2: 配置连接 Profile"

# 删除旧的测试配置
rm -rf ~/.ftpcli/test-config.yaml

# 添加 SFTP 配置
print_test "添加 SFTP profile"
./ftpcli profile add sftp-test \
    --protocol sftp \
    --host localhost \
    --port 2222 \
    --user testuser \
    --auth-type password \
    --password testpass \
    > /dev/null 2>&1 && print_pass "SFTP profile 添加成功" || print_fail "SFTP profile 添加失败"

# 添加 FTP 配置
print_test "添加 FTP profile"
./ftpcli profile add ftp-test \
    --protocol ftp \
    --host localhost \
    --port 21 \
    --user testuser \
    --auth-type password \
    --password testpass \
    > /dev/null 2>&1 && print_pass "FTP profile 添加成功" || print_fail "FTP profile 添加失败"

# 列出配置
print_test "列出所有 profiles"
./ftpcli profile list && print_pass "Profile 列表成功" || print_fail "Profile 列表失败"

print_header "步骤 3: 准备测试文件"

# 创建测试目录
mkdir -p test-data
cd test-data

# 创建测试文件
echo "Hello, ftpcli!" > test-small.txt
echo "This is a test file for ftpcli." > test-medium.txt

# 创建大文件 (10MB)
dd if=/dev/zero of=test-large.bin bs=1M count=10 > /dev/null 2>&1

# 创建测试目录
mkdir -p test-dir
echo "File 1 content" > test-dir/file1.txt
echo "File 2 content" > test-dir/file2.txt
mkdir -p test-dir/subdir
echo "Subfile content" > test-dir/subdir/subfile.txt

print_pass "测试文件创建成功"

cd ..

print_header "步骤 4: 测试 SFTP 功能"

# 测试 SFTP 列表
print_test "SFTP: 列出目录"
./ftpcli -p sftp-test ls /upload > /dev/null 2>&1 && print_pass "SFTP ls 成功" || print_fail "SFTP ls 失败"

# 测试 SFTP 上传小文件
print_test "SFTP: 上传小文件"
./ftpcli -p sftp-test upload test-data/test-small.txt /upload/ > /dev/null 2>&1 && \
    print_pass "SFTP 小文件上传成功" || print_fail "SFTP 小文件上传失败"

# 测试 SFTP 下载文件
print_test "SFTP: 下载文件"
./ftpcli -p sftp-test download /upload/test-small.txt test-data/downloaded-small.txt > /dev/null 2>&1 && \
    print_pass "SFTP 文件下载成功" || print_fail "SFTP 文件下载失败"

# 验证文件内容
if diff test-data/test-small.txt test-data/downloaded-small.txt > /dev/null 2>&1; then
    print_pass "SFTP 下载文件内容正确"
else
    print_fail "SFTP 下载文件内容不匹配"
fi

# 测试 SFTP 递归上传
print_test "SFTP: 递归上传目录"
./ftpcli -p sftp-test upload -r test-data/test-dir /upload/test-dir > /dev/null 2>&1 && \
    print_pass "SFTP 递归上传成功" || print_fail "SFTP 递归上传失败"

# 测试 SFTP 递归下载
print_test "SFTP: 递归下载目录"
./ftpcli -p sftp-test download -r /upload/test-dir test-data/downloaded-dir > /dev/null 2>&1 && \
    print_pass "SFTP 递归下载成功" || print_fail "SFTP 递归下载失败"

# 测试带校验和的上传
print_test "SFTP: 上传文件（带 MD5 校验）"
./ftpcli -p sftp-test upload --checksum test-data/test-medium.txt /upload/checksum-test.txt > /dev/null 2>&1 && \
    print_pass "SFTP 校验和上传成功" || print_fail "SFTP 校验和上传失败"

# 测试速率限制（大文件）
print_test "SFTP: 上传大文件（速率限制 1M）"
./ftpcli -p sftp-test upload --rate-limit 1M test-data/test-large.bin /upload/large.bin > /dev/null 2>&1 && \
    print_pass "SFTP 速率限制上传成功" || print_fail "SFTP 速率限制上传失败"

print_header "步骤 5: 测试 FTP 功能"

# 测试 FTP 列表
print_test "FTP: 列出目录"
./ftpcli -p ftp-test ls / > /dev/null 2>&1 && print_pass "FTP ls 成功" || print_fail "FTP ls 失败"

# 测试 FTP 上传
print_test "FTP: 上传文件"
./ftpcli -p ftp-test upload test-data/test-small.txt /test-small.txt > /dev/null 2>&1 && \
    print_pass "FTP 上传成功" || print_fail "FTP 上传失败"

# 测试 FTP 下载
print_test "FTP: 下载文件"
./ftpcli -p ftp-test download /test-small.txt test-data/ftp-downloaded.txt > /dev/null 2>&1 && \
    print_pass "FTP 下载成功" || print_fail "FTP 下载失败"

# 测试 FTP 详细列表
print_test "FTP: 详细列表 (ls -lh)"
./ftpcli -p ftp-test ls -lh / > /dev/null 2>&1 && print_pass "FTP ls -lh 成功" || print_fail "FTP ls -lh 失败"

print_header "步骤 6: 测试错误处理"

# 测试不存在的文件
print_test "错误处理: 下载不存在的文件"
./ftpcli -p sftp-test download /upload/nonexistent.txt test-data/error-test.txt > /dev/null 2>&1 && \
    print_fail "应该失败但成功了" || print_pass "正确处理文件不存在错误"

# 测试无效的 profile
print_test "错误处理: 使用不存在的 profile"
./ftpcli -p nonexistent-profile ls / > /dev/null 2>&1 && \
    print_fail "应该失败但成功了" || print_pass "正确处理 profile 不存在错误"

print_header "步骤 7: 清理测试环境"

echo "清理测试文件..."
rm -rf test-data

echo "停止 Docker 容器..."
docker-compose down > /dev/null 2>&1

print_pass "清理完成"

print_header "测试结果汇总"

echo ""
echo -e "${GREEN}通过: $TESTS_PASSED${NC}"
echo -e "${RED}失败: $TESTS_FAILED${NC}"
echo -e "总计: $((TESTS_PASSED + TESTS_FAILED))"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ 所有测试通过！${NC}"
    exit 0
else
    echo -e "${RED}✗ 有测试失败，请检查输出${NC}"
    exit 1
fi
