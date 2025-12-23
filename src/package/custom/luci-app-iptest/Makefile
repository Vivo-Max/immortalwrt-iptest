
# src/package/custom/luci-app-iptest/Makefile
include $(TOPDIR)/rules.mk

PKG_NAME:=luci-app-iptest
PKG_VERSION:=1.0
PKG_RELEASE:=1

include $(INCLUDE_DIR)/package.mk

define Package/luci-app-iptest
	SECTION:=luci
	CATEGORY:=LuCI
	SUBMENU:=3. Applications
	TITLE:=LuCI support for IP Test
	# 依赖 iptest 二进制包
	DEPENDS:=+iptest +luci-base +luci-compat +po2lmo
	PKGARCH:=all
endef

define Package/luci-app-iptest/description
	IP Test Tool Web Interface (with multi-language support).
endef

# LuCI app 不需要编译步骤
define Build/Prepare
endef

define Build/Compile
endef

# 核心：将 LuCI 文件复制到目标文件系统（$(1)）的正确位置
define Package/luci-app-iptest/install
	# --- 1. 安装 LuCI 脚本文件 (Controller, Model, View) ---
	
	# 安装 Controller
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/controller
	$(CP) ./files/luasrc/controller/* $(1)/usr/lib/lua/luci/controller/
	
	# 安装 Model/CBI
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/model/cbi
	$(CP) ./files/luasrc/model/cbi/* $(1)/usr/lib/lua/luci/model/cbi/
	
	# 安装 View
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/view/iptest
	$(CP) ./files/luasrc/view/iptest/* $(1)/usr/lib/lua/luci/view/iptest/
	
	
	# --- 2. 安装配置文件 ---
	
	$(INSTALL_DIR) $(1)/etc/config
	# 配置文件路径：files/root/etc/config/iptest
	$(INSTALL_CONF) ./files/root/etc/config/iptest $(1)/etc/config/iptest
	
	
	# --- 3. 手动处理并安装 i18n 翻译文件 (.po -> .lmo) ---
	# 原始版本中您使用了手动转换，这里恢复该逻辑以确保兼容性
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/i18n
	for po in ./po/*.po; do \
		if [ -f "$$po" ]; then \
			mo=$$(basename $$po .po).mo; \
			$(STAGING_DIR_HOST)/bin/po2lmo $$po $(1)/usr/lib/lua/luci/i18n/$$mo; \
		fi; \
	done
endef

$(eval $(call BuildPackage,luci-app-iptest))
