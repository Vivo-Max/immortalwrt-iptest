include $(TOPDIR)/rules.mk

PKG_NAME:=luci-app-iptest
PKG_VERSION:=1.0
PKG_RELEASE:=1

# 继承基础定义，这里使用 package.mk 或 luci.mk 都可以，
# 但手动定义安装步骤更灵活。
include $(INCLUDE_DIR)/package.mk

# 确保 po2lmo 可用，用于处理翻译文件
# 如果 OpenWrt/ImmortalWrt 版本较新，可以依赖 luci.mk 或 po2lmo
# 我们在这里依赖基础包和 iptest
define Package/luci-app-iptest
	SECTION:=luci
	CATEGORY:=LuCI
	SUBMENU:=3. Applications
	TITLE:=LuCI support for IP Test
	# 关键依赖：必须依赖 iptest 二进制包
	DEPENDS:=+iptest +luci-base +luci-compat +po2lmo
	PKGARCH:=all
endef

define Package/luci-app-iptest/description
	IP Test Tool Web Interface (with multi-language support).
endef

# LuCI 应用程序通常不需要编译，所以这两个为空
define Build/Prepare
endef

define Build/Compile
endef

# 核心：将 LuCI 文件复制到目标文件系统（$(1)）的正确位置
define Package/luci-app-iptest/install
	# --- 1. 安装 LuCI 脚本文件 (Controller, Model, View) ---
	
	# 检查文件结构: ./files/luasrc/controller/
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/controller
	$(CP) ./files/luasrc/controller/* $(1)/usr/lib/lua/luci/controller/
	
	# 检查文件结构: ./files/luasrc/model/cbi/
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/model/cbi
	$(CP) ./files/luasrc/model/cbi/* $(1)/usr/lib/lua/luci/model/cbi/
	
	# 检查文件结构: ./files/luasrc/view/iptest/
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/view/iptest
	$(CP) ./files/luasrc/view/iptest/* $(1)/usr/lib/lua/luci/view/iptest/
	
	
	# --- 2. 安装配置文件 ---
	
	# 检查文件结构: ./files/root/etc/config/iptest
	$(INSTALL_DIR) $(1)/etc/config
	$(INSTALL_CONF) ./files/root/etc/config/iptest $(1)/etc/config/iptest
	
	
	# --- 3. 自动处理和安装翻译文件 (.po -> .lmo) ---
	# OpenWrt/ImmortalWrt 的构建系统会自动处理 'po' 目录下的翻译文件。
	# 如果没有自动处理，您可以添加手动步骤，但通常不需要。
	# 我们在此依赖构建系统的默认行为。
endef

$(eval $(call BuildPackage,luci-app-iptest))
