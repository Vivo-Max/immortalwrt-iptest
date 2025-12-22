include $(TOPDIR)/rules.mk

# 软件包基础信息
PKG_NAME:=iptest
PKG_VERSION:=1.2
PKG_RELEASE:=1
PKG_LICENSE:=MIT

# 设置编译目录
PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)

# 依赖 Golang 环境
PKG_BUILD_DEPENDS:=golang/host

# Go 编译配置
# 注意：GO_PKG 必须与你 go.mod 文件中的 module 名称完全一致
GO_PKG:=iptest
GO_PKG_BUILD_BIN:=iptest
GO_PKG_ENABLE_CGO:=0

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

# 菜单显示定义
define Package/iptest
  SECTION:=net
  CATEGORY:=Network
  TITLE:=IP Test Tool
  DEPENDS:=$(GO_ARCH_DEPENDS)
endef

define Package/iptest/description
  A simple IP testing tool written in Go, built from local source.
endef

# --- 关键步骤：本地源码准备 ---
# 既然是 Actions 运行，直接将 Makefile 目录下的 go.mod, go.sum, iptest.go 拷贝到编译目录
define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
	$(CP) ./*.go $(PKG_BUILD_DIR)/
	$(CP) ./go.* $(PKG_BUILD_DIR)/
endef

# 安装生成的二进制文件到系统 /usr/bin/
define Package/iptest/install
	$(call GoPackage/Package/Install/Bin,$(1))
endef

$(eval $(call GoPackage,iptest))
$(eval $(call BuildPackage,iptest))
