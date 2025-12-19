#!/bin/bash

# 源代码文件名
SOURCE_FILE="iptest.go"

# 检查源代码是否存在
if [ ! -f "$SOURCE_FILE" ]; then
    echo "错误: 找不到 $SOURCE_FILE"
    exit 1
fi

# 确保依赖库已下载 (golang.org/x/net/proxy)
if [ ! -f "go.mod" ]; then
    echo "正在初始化 Go 模块..."
    go mod init iptest
    go mod tidy
fi

echo "开始编译多架构二进制文件..."

# 编译 x86_64 (用于 64位 PC 或 软路由)
echo "正在编译: iptest_x86_64..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o iptest_x86_64 $SOURCE_FILE

# 编译 ARM64 (用于 aarch64, 树莓派4, 现代 ARM 路由)
echo "正在编译: iptest_arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o iptest_arm64 $SOURCE_FILE

# 编译 ARMv7 (用于 32位 ARM, 如 K3, 昂达等旧设备)
echo "正在编译: iptest_armv7..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags="-s -w" -o iptest_armv7 $SOURCE_FILE

echo "---------------------------------------"
echo "编译完成！当前目录文件列表："
ls -lh iptest_*