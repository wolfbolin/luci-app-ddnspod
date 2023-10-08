module("luci.controller.ddnspod",package.seeall)
function index()
	entry({"admin", "services", "ddnspod"},cbi("wiolfi/ddnspod"),_("DDNS"),1)
end
