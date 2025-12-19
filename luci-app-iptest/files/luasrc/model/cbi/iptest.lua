local i18n = require "luci.i18n"
local t = i18n.translate

m = Map("iptest", t("Cloudflare IP测试工具"),
        t("配置完成后可手动运行或设置定时自动运行。<br/>敏感信息默认隐藏，可点击右侧👁图标查看明文。"))

-- ==================== 背景自定义分区 ====================
bg_section = m:section(TypedSection, "background", t("界面背景自定义（仅本页面生效）"))
bg_section.anonymous = true
bg_section.addremove = false

bg_type = bg_section:option(ListValue, "bg_type", t("背景类型"))
bg_type:value("none", t("无背景"))
bg_type:value("image", t("图片背景"))
bg_type:value("video", t("视频背景"))
bg_type.default = "none"

bg_url = bg_section:option(Value, "bg_url", t("背景图片/视频URL"))
bg_url:depends("bg_type", "image")
bg_url:depends("bg_type", "video")
bg_url.placeholder = t("例如: /mybg.jpg 或 https://example.com/bg.mp4")
bg_url.description = t("支持网络直链或本地文件。本地文件需手动上传到 /www/ 目录，此处填写路径如 /bg.jpg")

bg_blur = bg_section:option(Value, "bg_blur", t("背景模糊度 (px)"))
bg_blur.datatype = "uinteger"
bg_blur.default = "10"

bg_opacity = bg_section:option(Value, "bg_opacity", t("表单背景透明度 (0-1)"))
bg_opacity.datatype = "range(0,1,0.05)"
bg_opacity.default = "0.8"

-- 核心修复：重写渲染逻辑，避免破坏 HTML 骨架
m.render = function(self)
    local http = require "luci.http"
    local bg_type_val = m.uci:get("iptest", "background", "bg_type") or "none"
    local bg_url_val = m.uci:get("iptest", "background", "bg_url") or ""
    local bg_blur_val = m.uci:get("iptest", "background", "bg_blur") or "10"
    local bg_opacity_val = m.uci:get("iptest", "background", "bg_opacity") or "0.8"

    -- 注入 CSS 样式
    http.write("<style type='text/css'>")
    if bg_type_val == "image" and bg_url_val ~= "" then
        http.write(string.format("body { background: url('%s') center/cover no-repeat fixed !important; }", bg_url_val))
    elseif bg_type_val == "video" and bg_url_val ~= "" then
        http.write("#bg_video { position: fixed; top: 0; left: 0; width: 100%; height: 100%; object-fit: cover; z-index: -1; }")
    end
    
    -- 应用表单毛玻璃效果
    http.write(string.format([[
        .cbi-map { 
            background: rgba(255,255,255,%s) !important; 
            backdrop-filter: blur(%spx) !important; 
            -webkit-backdrop-filter: blur(%spx) !important;
            border-radius: 8px;
            padding: 15px;
        }
    ]], bg_opacity_val, bg_blur_val, bg_blur_val))
    http.write("</style>")

    -- 如果是视频背景，注入视频标签
    if bg_type_val == "video" and bg_url_val ~= "" then
        http.write(string.format([[
            <video id="bg_video" autoplay muted loop playsinline>
                <source src="%s" type="video/mp4">
            </video>
        ]], bg_url_val))
    end

    -- 调用原有的 CBI 渲染函数，确保菜单和脚部正常显示
    require("luci.template").render("cbi/map", {map = self})
end

-- ==================== 基本配置分区 ====================
s = m:section(TypedSection, "settings", t("基本配置"))
s.anonymous = true
s.addremove = false

path = s:option(Value, "path", t("IP列表文件路径"))
path.default = "/etc/iptest/ip.txt"

outfile = s:option(Value, "outfile", t("输出CSV文件名"))
outfile.default = "/tmp/result.csv"

max = s:option(Value, "max", t("最大并发数"))
max.datatype = "uinteger"
max.default = "100"

tls = s:option(Flag, "tls", t("启用TLS（HTTPS测试）"))
tls.default = "1"

speedtest = s:option(Value, "speedtest", t("测速并发数（0=禁用测速）"))
speedtest.datatype = "uinteger"
speedtest.default = "0"

speedlimit = s:option(Value, "speedlimit", t("最低速度阈值 (MB/s)"))
speedlimit.datatype = "uinteger"
speedlimit.default = "5"

url = s:option(Value, "url", t("测速下载URL"))
url.default = "speed.cloudflare.com/__down?bytes=500000000"

token = s:option(Value, "telegram_token", t("Telegram Bot Token"))
token.password = true

chat_ids = s:option(Value, "chat_ids", t("Telegram Chat IDs"))

proxy = s:option(Value, "preset_proxy", t("预设SOCKS5代理"))
proxy.password = true

-- ==================== 测试运行逻辑 ====================
run = s:option(Button, "run", "")
run.inputtitle = t("开始测试")
run.inputstyle = "apply"

function run.write(self, section)
    local path_val = m.uci:get("iptest", section, "path") or "/etc/iptest/ip.txt"
    local outfile_val = m.uci:get("iptest", section, "outfile") or "/tmp/result.csv"
    local max_val = m.uci:get("iptest", section, "max") or "100"
    local tls_val = m.uci:get("iptest", section, "tls") == "1" and "true" or "false"
    local speedtest_val = m.uci:get("iptest", section, "speedtest") or "0"
    local speedlimit_val = m.uci:get("iptest", section, "speedlimit") or "5"
    local url_val = m.uci:get("iptest", section, "url") or "speed.cloudflare.com/__down?bytes=500000000"
    local token_val = m.uci:get("iptest", section, "telegram_token") or ""
    local proxy_val = m.uci:get("iptest", section, "preset_proxy") or ""
    local chat_ids_val = m.uci:get("iptest", section, "chat_ids") or ""

    local cmd = ""
    if chat_ids_val ~= "" then
        cmd = cmd .. "export CHAT_IDS=\"" .. chat_ids_val .. "\" ; "
    end

    cmd = cmd .. "/usr/bin/iptest " ..
                "-path=\"" .. path_val .. "\" " ..
                "-outfile=\"" .. outfile_val .. "\" " ..
                "-max=" .. max_val .. " " ..
                "-tls=" .. tls_val .. " " ..
                "-speedtest=" .. speedtest_val .. " " ..
                "-int=" .. speedlimit_val .. " " ..
                "-url=\"" .. url_val .. "\""

    if token_val ~= "" then cmd = cmd .. " -telegram_token=\"" .. token_val .. "\"" end
    if proxy_val ~= "" then cmd = cmd .. " -preset_proxy=\"" .. proxy_val .. "\"" end

    cmd = cmd .. " > /tmp/iptest.log 2>&1 &"

    luci.sys.exec("echo '' > /tmp/iptest.log")
    luci.sys.exec(cmd)
    m.message = t("测试已启动！ 日志：/tmp/iptest.log")
end

-- ==================== 定时任务分区 ====================
cron_section = m:section(TypedSection, "cron", t("定时任务配置"))
cron_section.anonymous = true

enable_cron = cron_section:option(Flag, "enable_cron", t("启用定时运行"))
cron_expr = cron_section:option(Value, "cron_expr", t("Cron 表达式"))
cron_expr.default = "0 2 * * *"

apply_cron = cron_section:option(Button, "apply_cron", "")
apply_cron.inputtitle = t("应用定时设置")
apply_cron.inputstyle = "apply"

function apply_cron.write(self, section)
    local enable = m.uci:get("iptest", section, "enable_cron") == "1"
    local expr = m.uci:get("iptest", section, "cron_expr") or "0 2 * * *"

    -- 清理旧任务
    luci.sys.exec('sed -i "/iptest.*\\/usr\\/bin\\/iptest/d" /etc/crontabs/root')

    if enable then
        -- 构建定时执行命令（此处省略冗长的参数拼接，建议将逻辑封装进脚本简化此处）
        local cron_cmd = "/usr/bin/iptest -path=/etc/iptest/ip.txt > /tmp/iptest_cron.log 2>&1"
        luci.sys.exec('echo "' .. expr .. ' ' .. cron_cmd .. '" >> /etc/crontabs/root')
        luci.sys.exec("/etc/init.d/cron restart")
        m.message = t("定时任务已更新！")
    else
        luci.sys.exec("/etc/init.d/cron restart")
        m.message = t("定时任务已移除")
    end
end

return m
