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
# 指定Go主入口文件（基于您的iptest.go）
GO_PKG_BUILD_CMD:=./iptest.go

define Package/iptest
  SECTION:=net
  CATEGORY:=Network
  TITLE:=IP Test Tool
  DEPENDS:=$(GO_ARCH_DEPENDS) +luci-base +luci-compat +luci-mod-admin-full +po2lmo  # 添加LuCI依赖和po2lmo（用于po编译）
endef

define Package/iptest/description
  A simple IP testing tool written in Go, with LuCI interface.
endef

define Build/Prepare
	$(call Build/Prepare/Default)
	mkdir -p $(PKG_BUILD_DIR)
	# 复制Go相关文件到根目录
	$(CP) ./go.mod $(PKG_BUILD_DIR)/
	$(CP) ./go.sum $(PKG_BUILD_DIR)/
	$(CP) ./iptest.go $(PKG_BUILD_DIR)/
	# 复制luci-app-iptest子目录到构建目录
	$(CP) -r ./luci-app-iptest $(PKG_BUILD_DIR)/
endef

define Build/Compile
	# 先编译Go二进制
	$(call GoPackage/Build/Compile)
	# 运行po/Makefile编译po文件（生成mo文件）
	$(MAKE) -C $(PKG_BUILD_DIR)/luci-app-iptest/files/po
endef

define Package/iptest/install
	# 安装Go二进制（默认到/usr/bin/iptest）
	$(call GoPackage/Package/Install/Bin,$(1))
	# 安装LuCI控制器
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/controller
	$(CP) $(PKG_BUILD_DIR)/luci-app-iptest/files/luasrc/controller/iptest.lua $(1)/usr/lib/lua/luci/controller/
	# 安装LuCI CBI模型
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/model/cbi
	$(CP) $(PKG_BUILD_DIR)/luci-app-iptest/files/luasrc/model/cbi/iptest.lua $(1)/usr/lib/lua/luci/model/cbi/
	# 安装LuCI视图
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/view/iptest
	$(CP) $(PKG_BUILD_DIR)/luci-app-iptest/files/luasrc/view/iptest/log.htm $(1)/usr/lib/lua/luci/view/iptest/
	# 安装配置文件
	$(INSTALL_DIR) $(1)/etc/config
	$(CP) $(PKG_BUILD_DIR)/luci-app-iptest/files/root/etc/config/iptest $(1)/etc/config/
	# 安装i18n mo文件（假设po/Makefile生成在po目录下）
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/i18n
	$(CP) $(PKG_BUILD_DIR)/luci-app-iptest/files/po/*.mo $(1)/usr/lib/lua/luci/i18n/
endef

$(eval $(call GoBinPackage,iptest))
$(eval $(call BuildPackage,iptest))
