//go:build !linux

package arp

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

// StartARPListener listens for ARP packets and invokes the provided callback function.
func StartARPListener(callback func(event ARPEvent)) error {
	// Find all network devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return fmt.Errorf("failed to find network devices: %w", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("no network devices found")
	}

	// Use the first available device (you can modify this to select a specific device)
	device := devices[0].Name
	handle, err := pcap.OpenLive(device, 1600, true, pcap.BlockForever)
	if err != nil {
		return fmt.Errorf("failed to open device %s: %w", device, err)
	}
	defer handle.Close()

	// Set a filter to capture only ARP packets
	err = handle.SetBPFFilter("arp")
	if err != nil {
		return fmt.Errorf("failed to set BPF filter: %w", err)
	}

	// Use gopacket to process packets
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		arpLayer := packet.Layer(layers.LayerTypeARP)
		if arpLayer == nil {
			continue
		}

		arpPacket, _ := arpLayer.(*layers.ARP)
		event := ARPEvent{
			SenderIP:  fmt.Sprintf("%v", arpPacket.SourceProtAddress),
			SenderMAC: fmt.Sprintf("%v", arpPacket.SourceHwAddress),
			TargetIP:  fmt.Sprintf("%v", arpPacket.DstProtAddress),
			TargetMAC: fmt.Sprintf("%v", arpPacket.DstHwAddress),
		}

		// Determine the ARP operation type
		switch arpPacket.Operation {
		case layers.ARPRequest:
			event.Type = "request"
		case layers.ARPReply:
			event.Type = "reply"
		default:
			event.Type = "unknown"
		}

		// Invoke the callback with the ARP event
		callback(event)
	}
	return nil
}
