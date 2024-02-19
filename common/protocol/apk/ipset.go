//go:build linux

package apk

import (
	"github.com/nadoo/ipset"
)

var (
	ipsetTable = "dyn_v2ray_r2local"
)

func init() {
	// must call Init first
	if err := ipset.Init(); err != nil {
		panic("error in ipset Init: %s", err)
	}

	ipset.Create(ipsetTable)
}

func (h *SniffHeader) AddToIPset(ip string, timeout int) error {
	return ipset.Add(ipsetTable, ip, ipset.OptTimeout(timeout))
}
