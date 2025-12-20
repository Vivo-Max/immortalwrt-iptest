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
  # DEPENDS:= ... （如果需要依赖，如 +libpthread）
endef

define Package/iptest/description
  A simple IP testing tool written in Go.
endef

define Package/iptest/install
	\( (call GoPackage/Package/Install/Bin, \)(1))
endef

GO_PKG:=iptest

define Build/Compile
	$(call GoPackage/Build/Compile)
endef

$(eval $(call GoPackage,iptest))
$(eval $(call BuildPackage,iptest))
