include $(TOPDIR)/rules.mk

PKG_NAME:=luci-app-iptest
PKG_VERSION:=1.0
PKG_RELEASE:=1

# 继承 LuCI 宏，它会自动处理 po 语言文件
include $(TOPDIR)/feeds/luci/luci.mk

# 定义软件包信息
define Package/luci-app-iptest
	SECTION:=luci
	CATEGORY:=LuCI
	SUBMENU:=3. Applications
	TITLE:=LuCI support for IP Test
	# 关键：添加对 iptest 二进制包的依赖
	DEPENDS:=+iptest +luci-compat +luci-base +po2lmo
	PKGARCH:=all
endef

define Package/luci-app-iptest/description
	IP Test Tool Web Interface (with multi-language support).
endef

# 由于 LuCI 应用程序的源码都在 ./root/ 或 ./files/ 目录下
# 并且格式规范，可以简化安装步骤，直接按惯例复制文件结构。
define Package/luci-app-iptest/install
	# 严格按照 LuCI 文件结构安装 Controller, Model, View
	
	# 安装 Controller
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/controller
	$(CP) ./files/luasrc/controller/* $(1)/usr/lib/lua/luci/controller/
	
	# 安装 Model/CBI
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/model/cbi
	$(CP) ./files/luasrc/model/cbi/* $(1)/usr/lib/lua/luci/model/cbi/
	
	# 安装 View (注意：view 目录下通常是以 app 命名的子目录)
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/view/iptest
	$(CP) ./files/luasrc/view/iptest/* $(1)/usr/lib/lua/luci/view/iptest/
	
	# 安装配置文件
	$(INSTALL_DIR) $(1)/etc/config
	# 确保 ./files/root/etc/config/iptest 存在并使用 INSTALL_CONF
	$(INSTALL_CONF) ./files/root/etc/config/iptest $(1)/etc/config/iptest
	
	# 【注】i18n 翻译文件（po 目录）通常会被 make 自动处理和安装到 /usr/lib/lua/luci/i18n/
endef

$(eval $(call BuildPackage,luci-app-iptest))
