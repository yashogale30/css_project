package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const protocolID = "/p2p-chat/1.0.0"

func main() {
	ctx := context.Background()

	// Create libp2p host
	// creates crypto pair, secure connection,creates peer id, opens connections (tcp…)
	host, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Peer ID:", host.ID())

	// prints listening addr
	for _, a := range host.Addrs() {
		fmt.Printf("Listening on %s/p2p/%s\n", a, host.ID())
	}

	//recieving msgs
	host.SetStreamHandler(protocolID, func(s network.Stream) {
		fmt.Println("\nIncoming stream from:", s.Conn().RemotePeer())

		reader := bufio.NewReader(s)
		for {
			msg, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			fmt.Print("Peer: ", msg)
		}
	})

	//manually connecting peer, doing coz theres a issue comming in mdns in my local machine, can skip this later on
	fmt.Println("\nEnter peer multiaddress (or press Enter to skip):")

	var input string
	fmt.Scanln(&input)

	if input != "" {
		maddr, err := multiaddr.NewMultiaddr(input)
		if err != nil {
			log.Fatal(err)
		}

		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Fatal(err)
		}

		err = host.Connect(ctx, *info)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Connected to peer:", info.ID)
	}

	//send msg
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			text := scanner.Text()

			for _, p := range host.Network().Peers() {
				stream, err := host.NewStream(ctx, p, protocolID)
				if err != nil {
					continue
				}

				writer := bufio.NewWriter(stream)
				writer.WriteString(text + "\n")
				writer.Flush()
			}
		}
	}()

	//start mdns
	err = setupMDNS(ctx, host)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("mDNS service started")

	// Keep node alive
	select {}
}
