package arp

const (
	NEW int = iota
	DELETE
)

// ARPEvent represents an ARP event with its type and details.
type ARPEvent struct {
	Type      int // "request" or "reply"
	SenderIP  string
	SenderMAC string
}
