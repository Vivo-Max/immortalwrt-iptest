include $(TOPDIR)/rules.mk

PKG_NAME:=luci-app-iptest
PKG_VERSION:=1.0
PKG_RELEASE:=1

PKG_LICENSE:=GPL-2.0  # 根据实际许可证调整，如果不需要可移除

LUCI_TITLE:=LuCI support for IP Test
LUCI_DEPENDS:=+iptest
LUCI_PKGARCH:=all

include $(TOPDIR)/feeds/luci/luci.mk  # 引入 luci.mk 以标准化 LuCI 插件

define Package/$(PKG_NAME)/description
  Cloudflare IP测试工具 Web 界面（多语言支持）
endef

# =========================================================
# 准备（默认为空）
# =========================================================
define Build/Prepare
	$(call Build/Prepare/Default)
endef

# =========================================================
# 编译：处理 po 文件转换为 lmo
# =========================================================
define Build/Compile
	$(call Build/Compile/Default)
	# 查找所有 .po 文件并转换为 .lmo（支持多语言）
	for po in $(wildcard ./po/*/*.po); do \
		$(STAGING_DIR_HOST)/bin/po2lmo $$po $${po%.po}.lmo; \
	done
endef

# =========================================================
# 安装
# =========================================================
define Package/$(PKG_NAME)/install
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci
	$(CP) ./files/luasrc/* $(1)/usr/lib/lua/luci/
	# 安装转换后的 .lmo 文件到 i18n 目录
	$(INSTALL_DIR) $(1)/usr/lib/lua/luci/i18n
	$(CP) ./*.lmo $(1)/usr/lib/lua/luci/i18n/
	# 如果你有配置文件，也可以在这里安装
	# $(INSTALL_DIR) $(1)/etc/config
	# $(INSTALL_CONF) ./files/root/etc/config/iptest $(1)/etc/config/iptest
endef

# =========================================================
# 注册包
# =========================================================
$(eval $(call BuildPackage,$(PKG_NAME)))
