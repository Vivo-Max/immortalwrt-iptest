include $(TOPDIR)/rules.mk



PKG_NAME:=luci-app-iptest

PKG_VERSION:=1.0

PKG_RELEASE:=1



# 继承基础定义

include $(INCLUDE_DIR)/package.mk



define Package/luci-app-iptest

  SECTION:=luci

  CATEGORY:=LuCI

  SUBMENU:=3. Applications

  TITLE:=LuCI support for IP Test

  DEPENDS:=+iptest

  PKGARCH:=all

endef



define Package/luci-app-iptest/description

  Cloudflare IP测试工具 Web 界面（多语言支持）

endef



# 核心：LuCI 插件不需要编译，只需要将文件拷贝到对应位置

# po 目录下的文件会被构建系统自动处理成 .lmo 并安装

define Build/Prepare

endef



define Build/Compile

endef



define Package/luci-app-iptest/install

$(INSTALL_DIR) $(1)/usr/lib/lua/luci

$(CP) ./files/luasrc/* $(1)/usr/lib/lua/luci/


# 如果你有配置文件，也可以在这里安装

# $(INSTALL_DIR) $(1)/etc/config

# $(INSTALL_CONF) ./files/root/etc/config/iptest $(1)/etc/config/iptest

endef



$(eval $(call BuildPackage,luci-app-iptest))
