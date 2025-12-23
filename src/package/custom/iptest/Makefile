# src/package/custom/iptest/Makefile
# Copyright (C) 2025 Vivo-Max
#
# This is a package for ImmortalWrt/OpenWrt
#
include $(TOPDIR)/rules.mk
include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

PKG_NAME:=iptest
PKG_VERSION:=1.2
PKG_RELEASE:=1
PKG_LICENSE:=MIT
PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)
PKG_BUILD_DEPENDS:=golang/host

GO_PKG:=github.com/Vivo-Max/immortalwrt-iptest/iptest
GO_PKG_ENABLE_CGO:=0
# 指定Go主入口文件
GO_PKG_BUILD_CMD:=./iptest.go

define Package/iptest
	SECTION:=net
	CATEGORY:=Network
	TITLE:=IP Test Tool
	# 仅保留 Go 架构依赖，移除了所有 LuCI 依赖
	DEPENDS:=$(GO_ARCH_DEPENDS)
endef

define Package/iptest/description
	A simple IP testing tool written in Go.
endef

define Build/Prepare
	$(call Build/Prepare/Default)
	mkdir -p $(PKG_BUILD_DIR)
	# 只复制Go相关文件到构建目录
	$(CP) ./go.mod $(PKG_BUILD_DIR)/
	$(CP) ./go.sum $(PKG_BUILD_DIR)/
	$(CP) ./iptest.go $(PKG_BUILD_DIR)/
	# ❗ 移除：$(CP) -r ./luci-app-iptest $(PKG_BUILD_DIR)/ ❗
endef

define Build/Compile
	# 只编译Go二进制
	$(call GoPackage/Build/Compile)
endef

define Package/iptest/install
	# 只安装Go二进制到 /usr/bin/iptest
	$(call GoPackage/Package/Install/Bin,$(1))
	# 移除所有 LuCI 相关的安装代码
endef

$(eval $(call GoBinPackage,iptest))
$(eval $(call BuildPackage,iptest))
