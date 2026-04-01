package main

import (
	
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
)

const protocolID = "/p2p-chat/1.0.0"

// ChatMessage represents both text and voice messages
type ChatMessage struct {
	Type    string `json:"type"`    // "text" or "audio"
	Payload string `json:"payload"` // text content OR base64 audio data
	Sender  string `json:"sender"`  // Peer ID
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Global channel to pass messages from libp2p to the WebSocket UI
var uiMessages = make(chan ChatMessage, 10)

func main() {
	ctx := context.Background()

	host, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}
	defer host.Close()

	fmt.Println("Your Peer ID:", host.ID())

	// Handle incoming p2p streams
	host.SetStreamHandler(protocolID, func(s network.Stream) {
		defer s.Close()
		var msg ChatMessage
		
		// Read the JSON data sent from the peer
		data, err := io.ReadAll(s)
		if err != nil {
			return
		}
		
		if err := json.Unmarshal(data, &msg); err == nil {
			// Send the received message to the UI channel
			uiMessages <- msg
		}
	})

	// Start mDNS (using your existing setupMDNS function)
	err = setupMDNS(ctx, host)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("mDNS discovery started")

	// Setup HTTP Server for UI and WebSocket
	http.Handle("/", http.FileServer(http.Dir("./static")))
	
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		// Goroutine to send messages FROM libp2p TO the UI
		go func() {
			for msg := range uiMessages {
				conn.WriteJSON(msg)
			}
		}()

		// Loop to read messages FROM the UI and send TO libp2p peers
		for {
			var msg ChatMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				break
			}
			
			msg.Sender = host.ID().String()

			// Broadcast to all connected peers
			msgBytes, _ := json.Marshal(msg)
			for _, peerID := range host.Network().Peers() {
				stream, err := host.NewStream(ctx, peerID, protocolID)
				if err != nil {
					continue
				}
				stream.Write(msgBytes)
				stream.Close()
			}
		}
	})

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}