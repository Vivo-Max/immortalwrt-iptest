#
# Copyright (C) 2025 Vivo-Max
#
# This is a package for ImmortalWrt/OpenWrt
#

# 引用 OpenWrt/ImmortalWrt 的规则和 Go 语言包模板
include $(TOPDIR)/rules.mk
include $(INCLUDE_DIR)/package.mk
# 引入 Go 语言包核心支持
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

# ----------------------------------------------------
# 软件包基础信息
# ----------------------------------------------------
PKG_NAME:=iptest
PKG_VERSION:=1.2
PKG_RELEASE:=1
PKG_LICENSE:=MIT

# 设置编译目录 (保留您的原设置)
PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)

# 依赖 Golang Host 环境
PKG_BUILD_DEPENDS:=golang/host

# ----------------------------------------------------
# Go 编译配置 (已修正)
# ----------------------------------------------------
# 修复 1: GO_PKG 必须与 go.mod 文件中的 module 名称完全一致，
# 且应使用 Go 逻辑路径 (假设 go.mod 已修改为 'github.com/Vivo-Max/immortalwrt-iptest/iptest')
GO_PKG:=github.com/Vivo-Max/immortalwrt-iptest/iptest

# 编译后的二进制文件名
GO_PKG_BUILD_BIN:=iptest

# 禁用 CGO (OpenWrt/Musl 环境下的标准做法)
GO_PKG_ENABLE_CGO:=0

# ----------------------------------------------------
# 菜单显示定义
# ----------------------------------------------------
define Package/iptest
  SECTION:=net
  CATEGORY:=Network
  TITLE:=IP Test Tool
  DEPENDS:=$(GO_ARCH_DEPENDS)
endef

define Package/iptest/description
  A simple IP testing tool written in Go, built from local source.
endef

# ----------------------------------------------------
# 源码准备 (已修正 - 移除 Build/Prepare)
# ----------------------------------------------------
# 修复 2: 移除自定义的 Build/Prepare 宏。
# GoPackage 宏会自动将当前目录下的所有文件 (包括 go.mod, go.sum, *.go)
# 复制到 PKG_BUILD_DIR，并为 Go 交叉编译环境做好准备。
# 自定义 Build/Prepare 容易造成冲突和路径错误。
# define Build/Prepare
# 	mkdir -p $(PKG_BUILD_DIR)
# 	$(CP) ./*.go $(PKG_BUILD_DIR)/
# 	$(CP) ./go.* $(PKG_BUILD_DIR)/
# endef


# ----------------------------------------------------
# 安装步骤
# ----------------------------------------------------
# 使用 GoPackage 提供的标准安装宏，它知道二进制文件 GO_PKG_BUILD_BIN 的位置。
define Package/iptest/install
	$(call GoPackage/Package/Install/Bin,$(1))
endef

# ----------------------------------------------------
# 调用 Go 包构建系统
# ----------------------------------------------------
$(eval $(call GoPackage,iptest))
$(eval $(call BuildPackage,iptest))
