local sys = require("luci.sys")
local uci = require("luci.model.uci")

local m=Map("ddnspod",translate("WIOLFI-DDNS")) -- 标题

-- 基础设置
s=m:section(TypedSection,"base",translate("Base"))
s.anonymous=true -- 不显示section名称

enable=s:option(Flag,"enable",translate("enable"))
enable.rmempty=false

-- 秘钥设置
l1=m:section(TypedSection, "secret", translate("Secret"), "在此添加秘钥")
l1.template = "cbi/tblsection"
l1.addremove = true  -- 允许删改
l1.anonymous = true  -- 启用匿名

l1:option(Value, "name", "秘钥ID")
l1:option(Value, "model", "鉴权模式")
l1:option(Value, "secret_id", "鉴权用户")
l1:option(Value, "secret_key", "鉴权秘钥")

-- 监听器设置
l2=m:section(TypedSection, "listener", translate("Listener"), "在此添加监听器")
l2.template = "cbi/tblsection"
l2.addremove = true  -- 允许删改
l2.anonymous = true  -- 启用匿名

o=l2:option(Value, "name", "监听器ID")

mod=l2:option(ListValue, "model", "监听模式")
mod:value("netlink", translate("netlink"))
mod:value("query", translate("query"))

et=l2:option(ListValue, "eth_type", "网络类型")
et:value("ipv4", "ipv4")
et:value("ipv6", "ipv6")

en=l2:option(ListValue, "eth_name", "网卡名称")
local cmd="ip link show | grep state"
for line in sys.exec(cmd):gmatch("[^\r\n]+") do
	local inf_name = line:match("^%d+: ([%w%.%-]+):")
	en:value(inf_name, inf_name)
end

-- 依赖库方案
-- local nw = require("luci.model.network").init()
-- local ifaces = nw:get_interfaces()
-- for _, iface in ipairs(ifaces) do
--     local row_name = iface:name()
-- 	en:value(row_name, row_name)
-- end

-- 配置文件方案
-- local c = require("luci.model.uci").cursor()
-- c:foreach("network", "interface",
--     function(section)
--         local row_name = section[".name"]
-- 		   eth_name:value(row_name, row_name)
--     end
-- )

-- ubus方案
-- local util = require("luci.util")
-- local ifaces = util.ubus("network.interface", "dump", { })
-- if ifaces then
-- 	for _, iface in ipairs(ifaces.interface) do
-- 		en:value(iface.interface, iface.interface)
-- 	end
-- end

-- 供应商设置
l3=m:section(TypedSection, "provider", translate("Provider"), "在此添加供应商")
l3.template = "cbi/tblsection"
l3.addremove = true  -- 允许删改
l3.anonymous = true  -- 启用匿名

l3:option(Value, "name", "供应商ID")

mod=l3:option(ListValue, "model", "解析方案")
mod:value("dnspod", translate("dnspod"))

l3:option(Value, "secret", "秘钥ID")
l3:option(Value, "sub_domain", "子域名")
l3:option(Value, "main_domain", "主域名")

-- 日志面板
l4=m:section(TypedSection,"base",translate("Update Log"))
l4.anonymous=true
local a="/var/log/tencentddns.log"
tvlog=l4:option(TextValue,"sylogtext")
tvlog.rows=16
tvlog.readonly="readonly"
tvlog.wrap="off"

function tvlog.cfgvalue(l4,l4)
	sylogtext=""
	if a and nixio.fs.access(a) then
		sylogtext=luci.sys.exec("tail -n 100 %s"%a)
	end
	return sylogtext
end


tvlog.write=function(l4,l4,l4)
end


local apply=luci.http.formvalue("cbi.apply")
if apply then
	io.popen("/etc/init.d/ddnspod restart")
end

local submit=luci.http.formvalue("cbi.submit")
if submit then
end

return m