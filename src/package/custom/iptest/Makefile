include $(TOPDIR)/rules.mk

PKG_NAME:=iptest
PKG_VERSION:=1.2
PKG_RELEASE:=1

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/iptest
  SECTION:=net
  CATEGORY:=Network
  TITLE:=iptest - IP Test Tool
  # DEPENDS:= +libpthread  # 如果有依赖，取消注释并调整
endef

define Package/iptest/description
  A simple IP testing tool written in Go.
endef

define Build/Prepare
	$(call Build/Prepare/Default)
	# 假设源代码在包目录的 src/ 下，复制到构建目录
	mkdir -p $(PKG_BUILD_DIR)
	cp -r $(PKG_BUILD_DIR)/../src/* $(PKG_BUILD_DIR)/ || true  # 调整为您的源路径
endef

define Build/Compile
	$(call GoPackage/Build/Compile)
endef

define Package/iptest/install
	$(call GoPackage/Package/Install/Bin,$(1))
endef

GO_PKG:=iptest  # 如果是 GitHub 仓库，改为 github.com/youruser/iptest

$(eval $(call GoPackage,iptest))
$(eval $(call BuildPackage,iptest))
