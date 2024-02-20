//go:build linux

package apk

import (
	"fmt"

	"github.com/nadoo/ipset"
	"github.com/v2fly/v2ray-core/v5/common/net"
)

var (
	ipsetTable = "dyn_v2ray_r2local"
)

func init() {
	// must call Init first
	if err := ipset.Init(); err != nil {
		panic("error in ipset Init:" + err.Error())
	}

	ipset.Create(ipsetTable)
}

func (h *SniffHeader) AddToIPset(addr net.Destination, timeout int) error {
	switch addr.Address.Family() {
	case net.AddressFamilyDomain:
		return fmt.Errorf("sniff.apk-download.found.a.domain:%v", addr.Address.Domain())
	case net.AddressFamilyIPv4:
		return ipset.Add(ipsetTable, addr.Address.IP().String(), ipset.OptTimeout(timeout))
	case net.AddressFamilyIPv6:
		fallthrough
	default:
		return fmt.Errorf("sniff.apk-download.found.unknown.ip:%v", addr.Address.String())
	}
}
