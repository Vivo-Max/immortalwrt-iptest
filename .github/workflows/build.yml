include $(TOPDIR)/rules.mk

PKG_NAME:=iptest
PKG_VERSION:=1.2        # 随便填个版本号
PKG_RELEASE:=1

include $(INCLUDE_DIR)/package.mk
include $(TOPDIR)/feeds/packages/lang/golang/golang-package.mk

define Package/iptest
  SECTION:=net
  CATEGORY:=Network
  TITLE:=iptest - IP Test Tool
  # 如果有依赖（如需要 libc 或其他包），在这里加，例如 DEPENDS:= +libpthread
endef

define Package/iptest/description
  A simple IP testing tool written in Go.
endef

# 模块名必须和 go.mod 里的 module xxx 一致
GO_PKG:=iptest

# 如果你的程序有多个 binary 或特殊编译 flags，可以在这里加
# GO_PKG_BUILD_FLAGS:= -tags netgo

define Build/Compile
	$(call GoPackage/Build/Compile)
endef

$(eval $(call GoPackage,iptest))
$(eval $(call BuildPackage,iptest))
