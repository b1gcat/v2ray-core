package arp

// ARPEvent represents an ARP event with its type and details.
type ARPEvent struct {
	Type      string // "request" or "reply"
	SenderIP  string
	SenderMAC string
	TargetIP  string
	TargetMAC string
}
