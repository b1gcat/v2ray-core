//go:build !linux

package arp

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
)

// ClearARPCache clears the ARP cache based on the operating system
func ClearARPCache() error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("netsh", "interface", "ip", "delete", "arpcache")
	case "darwin":
		cmd = exec.Command("arp", "-a", "-d")
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to clear ARP cache: %w, output: %s", err, string(output))
	}
	return nil
}

// StartARPListener listens for ARP packets and invokes the provided callback function.
func StartARPListener(callback func(event ARPEvent)) error {
	if err := ClearARPCache(); err != nil {
		logrus.Errorf("ARP cache clearing failed: %w", err)
	}
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
		}

		// Determine the ARP operation type
		switch arpPacket.Operation {
		case layers.ARPReply:
		default:
			continue
		}
		event.Type = NEW
		// Invoke the callback with the ARP event
		callback(event)
	}
	return nil
}
