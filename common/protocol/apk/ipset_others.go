//go:build !linux

package apk

import (
	"fmt"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

func (h *SniffHeader) AddToIPset(addr net.Destination) error {
	switch addr.Address.Family() {
	case net.AddressFamilyDomain:
		return fmt.Errorf("sniff.apk-download.found.a.domain: %v", addr.Address.Domain())
	case net.AddressFamilyIPv4, net.AddressFamilyIPv6:
		return fmt.Errorf("sniff.apk-download.found.a.ip: %v", addr.Address.IP().String())
	default:
		return fmt.Errorf("sniff.apk-download.found.unknown.ip: %v", addr.Address.String())
	}
}
