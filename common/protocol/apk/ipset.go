//go:build linux

package apk

import (
	"context"
	"fmt"
	"time"

	"github.com/nadoo/ipset"
	"github.com/v2fly/v2ray-core/v5/common/net"

	net2 "net"
)

var (
	ipsetTable = "apkWhiteIP"
)

func init() {
	// must call Init first
	if err := ipset.Init(); err != nil {
		panic("error in ipset Init:" + err.Error())
	}
}

func (h *SniffHeader) AddToIPset(addr net.Destination) error {
	switch addr.Address.Family() {
	case net.AddressFamilyDomain:
		ips, err := lookUpHost(addr.Address.Domain())
		if err != nil {
			return fmt.Errorf("sniff.apk-download.found.a.domain: %v:%v",
				addr.Address.Domain(), err.Error())
		}
		for _, ip := range ips {
			ipset.Add(ipsetTable, ip, ipset.OptTimeout(86400))
		}

		return fmt.Errorf("sniff.apk-download.found.a.domain: %v insert IP %v",
			addr.Address.Domain(), ips)
	case net.AddressFamilyIPv4:
		ipset.Add(ipsetTable, addr.Address.IP().String(), ipset.OptTimeout(86400))
		return fmt.Errorf("success add %v", addr.Address.IP().String())
	case net.AddressFamilyIPv6:
		fallthrough
	default:
		return fmt.Errorf("sniff.apk-download.found.unknown.ip: %v", addr.Address.String())
	}
}

func lookUpHost(domain string) ([]string, error) {
	// 设置超时时间为3秒
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// 使用WithContext方法创建一个新的Resolver对象，并将其上下文设置为ctx
	resolver := net2.Resolver{
		PreferGo: true,
		Dial:     (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
	}
	return resolver.LookupHost(ctx, domain)
}
