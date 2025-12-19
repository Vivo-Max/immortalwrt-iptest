include $(TOPDIR)/rules.mk

PKG_NAME:=iptest
PKG_VERSION:=1.2
PKG_RELEASE:=1

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/iptest
  SECTION:=utils
  CATEGORY:=Utilities
  TITLE:=Cloudflare IP Tester
  DEPENDS:=$(GO_ARCH_DEPENDS) +ca-bundle
endef

# 核心修正：取消 GO_PKG 的自动推导，直接指定 build 路径
GO_PKG_BUILD_PKG:=.

define Build/Prepare
	mkdir -p $(PKG_BUILD_DIR)
	$(CP) ./src/* $(PKG_BUILD_DIR)/
endef

define Build/Compile
	( cd $(PKG_BUILD_DIR); \
		$(GO_PKG_VARS) \
		go build -v -o $(PKG_INSTALL_DIR)/usr/bin/iptest ./iptest.go \
	)
endef

define Package/iptest/install
	$(INSTALL_DIR) $(1)/usr/bin
	$(INSTALL_BIN) $(PKG_INSTALL_DIR)/usr/bin/iptest $(1)/usr/bin/iptest
endef

$(eval $(call BuildPackage,iptest))
