#!/bin/bash
# 使用 sftp 命令测试
sftp -P 2222 testuser@localhost << 'SFTP'
testpass
cd /upload
mkdir test-manual-dir
ls -la
bye
SFTP
