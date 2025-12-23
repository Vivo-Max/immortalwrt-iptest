include $(TOPDIR)/rules.mk

PKG_NAME:=iptest
PKG_VERSION:=1.2
PKG_RELEASE:=1

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)
PKG_BUILD_DEPENDS:=golang/host  # 依赖 host Go 工具
PKG_BUILD_PARALLEL:=1  # 启用并行构建

PKG_MAINTAINER:=ang <your@email>  # 可替换为您的信息
PKG_LICENSE:=GPL-2.0  # 根据实际许可证调整

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

# =========================================================
# Package 定义
# =========================================================
define Package/iptest
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=Cloudflare IP Tester
  DEPENDS:=$(GO_ARCH_DEPENDS) +ca-bundle
endef

define Package/iptest/description
  Cloudflare IP Tester written in Go
endef

# =========================================================
# Go 构建参数
# =========================================================
# Go module 根目录（如果 go.mod 有指定模块名，请替换；否则 . 也可）
GO_PKG:=.
# main 包所在路径（假设 main.go 在 cmd/iptest/main.go）
GO_PKG_BUILD_PKG:=./cmd/iptest
# 生成的二进制名称
GO_PKG_BUILD_BIN:=iptest

# =========================================================
# 准备源码（复制本地 src/ 目录下的源码）
# =========================================================
define Build/Prepare
	$(call Build/Prepare/Default)
	$(CP) ./src/* $(PKG_BUILD_DIR)/
endef

# =========================================================
# 安装
# =========================================================
define Package/iptest/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(GO_PKG_BUILD_BIN_DIR)/$(GO_PKG_BUILD_BIN) $(1)/usr/bin/iptest
endef

# =========================================================
# 注册包
# =========================================================
$(eval $(call GoBinPackage,iptest))
$(eval $(call BuildPackage,iptest))
