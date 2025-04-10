//go:build linux

package arp

import (
	"fmt"
	"log"
	"net"
	"unsafe"

	"golang.org/x/sys/unix"
)

// StartARPListener listens for ARP events and invokes the provided callback function.
func StartARPListener(callback func(event ARPEvent)) error {
	sock, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW, unix.NETLINK_ROUTE)
	if err != nil {
		return fmt.Errorf("failed to create netlink socket: %w", err)
	}
	defer unix.Close(sock)

	err = unix.Bind(sock, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: unix.RTMGRP_NEIGH,
	})
	if err != nil {
		return fmt.Errorf("failed to bind netlink socket: %w", err)
	}

	buf := make([]byte, 4096)
	for {
		n, _, err := unix.Recvfrom(sock, buf, 0)
		if err != nil {
			log.Printf("Failed to receive netlink message: %v", err)
			continue
		}

		msgs, err := unix.ParseNetlinkMessage(buf[:n])
		if err != nil {
			log.Printf("Failed to parse netlink message: %v", err)
			continue
		}

		for _, msg := range msgs {
			if msg.Header.Type == unix.RTM_NEWNEIGH || msg.Header.Type == unix.RTM_DELNEIGH || msg.Header.Type == unix.RTM_SETNEIGH {
				ndm := *(*unix.Ndmsg)(unsafe.Pointer(&msg.Data[0]))

				// Parse attributes to extract IP and MAC addresses
				attrs, err := unix.ParseNetlinkRouteAttr(&msg)
				if err != nil {
					log.Printf("Failed to parse netlink attributes: %v", err)
					continue
				}

				var senderIP, targetIP net.IP
				var senderMAC, targetMAC net.HardwareAddr
				for _, attr := range attrs {
					switch attr.Attr.Type {
					case unix.NDA_DST: // IP address
						senderIP = net.IP(attr.Value)
					case unix.NDA_LLADDR: // MAC address
						senderMAC = net.HardwareAddr(attr.Value)
					}
				}

				// Populate the ARPEvent struct
				event := ARPEvent{
					SenderIP:  senderIP.String(),
					SenderMAC: senderMAC.String(),
					TargetIP:  targetIP.String(),  // Target IP is not directly available in netlink
					TargetMAC: targetMAC.String(), // Target MAC is not directly available in netlink
				}

				switch msg.Header.Type {
				case unix.RTM_NEWNEIGH:
					event.Type = "add"
				case unix.RTM_DELNEIGH:
					event.Type = "delete"
				case unix.RTM_SETNEIGH:
					event.Type = "modify"
				}

				// Invoke the callback with the ARP event
				callback(event)
			}
		}
	}
}
