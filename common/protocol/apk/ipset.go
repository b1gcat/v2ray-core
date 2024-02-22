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

	ipset.Create(ipsetTable, ipset.OptTimeout(86400))
}

func (h *SniffHeader) AddToIPset(addr net.Destination) error {
	switch addr.Address.Family() {
	case net.AddressFamilyDomain:
		return fmt.Errorf("sniff.apk-download.found.a.domain: %v", addr.Address.Domain())
	case net.AddressFamilyIPv4:
		ipset.Add(ipsetTable, addr.Address.IP().String())
		return fmt.Errorf("success add %v", addr.Address.IP().String())
	case net.AddressFamilyIPv6:
		fallthrough
	default:
		return fmt.Errorf("sniff.apk-download.found.unknown.ip: %v", addr.Address.String())
	}
}
