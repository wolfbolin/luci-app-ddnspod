package main

import (
	"context"
	"ddnspod/config"
	"ddnspod/solution"
	"flag"
	"fmt"
	"os"
	"sync"
)

var configFilePath = flag.String("c", "/etc/config/ddnspod", "配置文件路径")

func init() {
	flag.Parse()
}

func main() {
	// 读取本地监控策略配置
	configMap, err := config.ParserConfig(*configFilePath)
	if err != nil {
		fmt.Printf("Parser config file error: %s\n", err)
		os.Exit(1)
	}

	// 优先创建鉴权对象
	secretBin := make(map[string]solution.Secret)
	for _, item := range configMap.Section {
		if item.Type == "secret" && item.Option["model"] == "dnspodv3" {
			secretBin[item.Key] = solution.NewDnsPodSecret(item.Option["secret_id"], item.Option["secret_key"])
		}
	}

	// 创建解析服务对象
	wg := sync.WaitGroup{}
	ctx := context.Background()
	dnsProvider := make(map[string]solution.Provider)
	for _, item := range configMap.Section {
		if item.Type == "provider" && item.Option["model"] == "dnspod" {
			mod, initErr := solution.NewDnsPodResolver(ctx, item.Option["main_domain"], item.Option["sub_domain"], secretBin[item.Option["secret"]])
			if initErr != nil {
				panic(initErr)
			}
			dnsProvider[item.Key] = mod

			wg.Add(1)
			go mod.StartMod(&wg)
			fmt.Printf("[MAIN] Create provider <%s> goroutine\n", item.Key)
		}
	}

	// 创建IP监听对象
	for _, item := range configMap.Section {
		if item.Type == "listener" && item.Option["model"] == "netlink" {
			mod, initErr := solution.NewLocalListener(ctx, item.Option["eth_name"], item.Option["eth_type"])
			if initErr != nil {
				panic(initErr)
			}
			for _, provider := range item.List["provider"] {
				fmt.Printf("[MAIN] Add ip pipe from [%s] to [%s]\n", item.Key, provider)
				dnsProvider[provider].UpdateIP(mod.IpUpdate())
			}

			wg.Add(1)
			go mod.StartMod(&wg)
			fmt.Printf("[MAIN] Create listener <%s> goroutine\n", item.Key)
		}
	}

	// 持续运行
	fmt.Println("[MAIN] Program running...")
	wg.Wait()
	fmt.Println("[MAIN] Program stopped running")
}
