package solution

import (
	"context"
	"ddnspod/util"
	"fmt"
	"github.com/vishvananda/netlink"
	"sync"
	"time"
)

type NetLinkListener struct {
	ethName    string
	ethType    string
	eventChan  chan netlink.AddrUpdate
	updateChan chan util.EthIP
}

func NewLocalListener(ctx context.Context, ethName, ethType string) (*NetLinkListener, error) {
	// 创建本地监听器对象
	r := NetLinkListener{
		ethName:    ethName,
		ethType:    ethType,
		eventChan:  make(chan netlink.AddrUpdate),
		updateChan: make(chan util.EthIP),
	}

	// 添加本地网络变化监听
	err := netlink.AddrSubscribe(r.eventChan, ctx.Done())
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *NetLinkListener) StartMod(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		fmt.Println("[NetLink] Goroutine exited")
	}()

	// 启动时主动完成一次刷写
	r.syncEthIP("")

	for e := range r.eventChan {
		// 当网卡出现新的IP时才触发变化
		if !e.NewAddr {
			continue
		}
		r.syncEthIP(e.LinkAddress.IP.String())
	}
}

func (r *NetLinkListener) IpUpdate() chan util.EthIP {
	return r.updateChan
}

func (r *NetLinkListener) syncEthIP(eventIP string) {
	// 获取本地网口信息
	ethIPs, err := util.GetNetEthIPs()
	if err != nil {
		fmt.Printf("[NetLink] Get eth ip error: %s", err.Error())
	}

	// 当变动网口信息匹配时更新
	for _, ethIP := range ethIPs {
		if ethIP.Name == r.ethName && ethIP.Type == r.ethType {
			if eventIP == "" || ethIP.IP == eventIP {
				timeNow := time.Now()
				fmt.Printf("[LOCAL] Observed IP change to %s @ %s\n", ethIP.IP, timeNow.Format("2006-01-02 15:04:05"))
				r.updateChan <- ethIP
				break
			}
		}
	}
}
