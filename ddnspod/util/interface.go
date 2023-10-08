package util

import (
	"fmt"
	"net"
)

// EthIP 本机网络
type EthIP struct {
	IP   string
	IDX  int
	MTU  int
	MAC  string
	Mask int
	Name string
	Type string
}

// GetNetEthIPs 获得所有网卡的ipv4和ipv6地址
func GetNetEthIPs() ([]EthIP, error) {
	// 获取所有本地网络接口信息
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	interfaceList := make([]EthIP, 0)
	for _, netIf := range netInterfaces {
		ifAddresses, err := netIf.Addrs()
		if err != nil {
			return nil, err
		}

		for _, ifAddress := range ifAddresses {
			ethIP, ok := ifAddress.(*net.IPNet)
			if !ok {
				return nil, fmt.Errorf("Interface type assertion error: *net.IPNet\n")
			}

			if !ethIP.IP.IsGlobalUnicast() {
				continue
			}
			mask, bits := ethIP.Mask.Size()
			ipType := "unknown"
			if bits == 128 {
				ipType = "ipv6"
			} else if bits == 32 {
				ipType = "ipv4"
			}

			inf := EthIP{
				IP:   ethIP.IP.String(),
				IDX:  netIf.Index,
				MTU:  netIf.MTU,
				MAC:  netIf.HardwareAddr.String(),
				Mask: mask,
				Name: netIf.Name,
				Type: ipType,
			}
			interfaceList = append(interfaceList, inf)
		}
	}
	return interfaceList, nil
}
