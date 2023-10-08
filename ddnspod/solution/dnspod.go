package solution

import (
	"context"
	"ddnspod/util"
	"fmt"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
	"sync"
	"time"
)

type DnsPodSecret struct {
	SecretId  string
	SecretKey string
}

func NewDnsPodSecret(secretId, secretKey string) *DnsPodSecret {
	return &DnsPodSecret{
		SecretId:  secretId,
		SecretKey: secretKey,
	}
}

func (d *DnsPodSecret) GetAuth() any {
	credential := common.NewCredential(
		d.SecretId,
		d.SecretKey,
	)
	return credential
}

type DnsPodResolver struct {
	updateChan chan util.EthIP
	mainDomain string
	subDomain  string
	ipCache    *util.EthIP
	client     *dnspod.Client
	ctx        context.Context
}

func NewDnsPodResolver(ctx context.Context, mainDomain, subDomain string, pass Secret) (*DnsPodResolver, error) {
	credential := pass.GetAuth().(*common.Credential)
	cpf := profile.NewClientProfile()
	client, err := dnspod.NewClient(credential, "", cpf)
	if err != nil {
		return nil, err
	}

	r := DnsPodResolver{
		updateChan: nil,
		mainDomain: mainDomain,
		subDomain:  subDomain,
		ipCache:    nil,
		client:     client,
		ctx:        ctx,
	}
	return &r, nil
}

func (r *DnsPodResolver) StartMod(wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
		fmt.Println("[NetLink] Goroutine exited")
	}()

	checkTicker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-r.ctx.Done():
			return
		case <-checkTicker.C:
			if r.ipCache == nil {
				continue
			}
		case ethIP := <-r.updateChan:
			if r.ipCache != nil && r.ipCache.IP == ethIP.IP {
				continue
			}
			r.ipCache = &ethIP
		}
		err := r.syncDNS()
		if err != nil {
			fmt.Printf("[DNSPOD] Update dns error: %s", err.Error())
		}
	}
}

func (r *DnsPodResolver) UpdateIP(msgEth chan util.EthIP) {
	r.updateChan = msgEth
}

func (r *DnsPodResolver) syncDNS() error {
	// 尝试获取用户域名下解析列表
	queryReq := dnspod.NewDescribeRecordFilterListRequest()
	queryReq.Domain = common.StringPtr(r.mainDomain)
	queryReq.SubDomain = common.StringPtr(r.subDomain)
	queryRes, err := r.client.DescribeRecordFilterList(queryReq)
	if e, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("Bad request : %s.[Code %s]\n", e.Message, e.Code)
	}
	if err != nil {
		return fmt.Errorf("Runtime error at DescribeRecordFilterList: %s\n", err.Error())
	}

	// 当且仅当存在唯一记录时更新
	if *queryRes.Response.RecordCountInfo.ListCount == 0 {
		return fmt.Errorf("Record %s.%s does not exist. Please add this record manually.\n", r.subDomain, r.mainDomain)
	}
	if *queryRes.Response.RecordCountInfo.ListCount >= 2 {
		return fmt.Errorf("Record %s.%s is not unique. Please delete excess records manually.\n", r.subDomain, r.mainDomain)
	}
	dnsCache := queryRes.Response.RecordList[0]

	// IP不一致时同步DNS数据
	if *dnsCache.Type == "A" && *dnsCache.Value == r.ipCache.IP {
		return nil
	}
	modifyReq := dnspod.NewModifyRecordRequest()
	modifyReq.Value = common.StringPtr(r.ipCache.IP)
	modifyReq.Domain = common.StringPtr(r.mainDomain)
	modifyReq.SubDomain = common.StringPtr(r.subDomain)
	modifyReq.RecordType = common.StringPtr("A")
	modifyReq.RecordLine = common.StringPtr(*dnsCache.Line)
	modifyReq.RecordId = common.Uint64Ptr(*dnsCache.RecordId)
	_, err = r.client.ModifyRecord(modifyReq)
	if e, ok := err.(*errors.TencentCloudSDKError); ok {
		return fmt.Errorf("Bad request : %s.[Code %s]\n", e.Message, e.Code)
	}
	if err != nil {
		return fmt.Errorf("Runtime error at ModifyRecord: %s\n", err.Error())
	}
	return nil
}
