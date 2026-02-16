package main

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const mdnsServiceName = "p2p-chat-mdns"

// Notifee is called when a peer is discovered
type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Println("Found peer via mDNS:", pi.ID)

	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Println("mDNS connect failed:", err)
	}
}

// Start mDNS discovery
func setupMDNS(ctx context.Context, h host.Host) error {
	notifee := &discoveryNotifee{h: h}
	service := mdns.NewMdnsService(h, mdnsServiceName, notifee)
	return service.Start()
}
