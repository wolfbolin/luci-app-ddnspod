package solution

import (
	"ddnspod/util"
	"sync"
)

type Secret interface {
	GetAuth() any
}

type Provider interface {
	StartMod(*sync.WaitGroup)
	UpdateIP(chan util.EthIP)
}

type Listener interface {
	StartMod(*sync.WaitGroup)
	IpUpdate() chan util.EthIP
}
