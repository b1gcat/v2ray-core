//go:build !linux

package apk

import "fmt"

func (h *SniffHeader) AddToIPset(ip string, timeout int) error {
	return fmt.Errorf("not support ipset module for %v/%v", ip, timeout)
}
