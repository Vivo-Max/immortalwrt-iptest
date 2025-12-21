include $(TOPDIR)/rules.mk

PKG_NAME:=iptest
PKG_VERSION:=1.2
PKG_RELEASE:=1

PKG_SOURCE_PROTO:=git
PKG_SOURCE_URL:=https://github.com/youruser/iptest.git
PKG_SOURCE_VERSION:=HEAD

PKG_BUILD_DIR:=$(BUILD_DIR)/$(PKG_NAME)-$(PKG_VERSION)

PKG_LICENSE:=MIT
PKG_LICENSE_FILES:=LICENSE

PKG_BUILD_DEPENDS:=golang/host

# Go module 路径（必须和 go.mod 的 module 完全一致）
GO_PKG:=github.com/youruser/iptest

# main.go 在仓库根目录
GO_PKG_BUILD_PKG:=.

include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk
include $(INCLUDE_DIR)/package.mk

define Package/iptest
  SECTION:=net
  CATEGORY:=Network
  TITLE:=IP Test Tool
endef

define Package/iptest/description
A simple IP testing tool written in Go.
endef

define Package/iptest/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(GO_PKG_BUILD_BIN_DIR)/iptest $(1)/usr/bin/
endef

$(eval $(call GoBinPackage,iptest))
$(eval $(call BuildPackage,iptest))
