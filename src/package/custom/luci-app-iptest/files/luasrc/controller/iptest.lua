module("luci.controller.iptest", package.seeall)

local i18n = require "luci.i18n"
local t = i18n.translate

function index()
    entry({"admin", "services", "iptest"}, cbi("iptest"), t("Cloudflare IP测试"), 100).dependent = true
    entry({"admin", "services", "iptest", "log"}, view("iptest/log"), t("运行日志"), 2)
end
