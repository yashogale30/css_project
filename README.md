#  LAN Mesh — Serverless P2P Communication

> Text and voice messaging over your local network. No internet. No servers. No accounts. Just devices on the same Wi-Fi.

---

## What is this?

Most chat apps secretly rely on a data center somewhere. Your message leaves your device, travels to a cloud server, and comes back down to the person sitting next to you. Cut the internet, and everything stops.

**LAN Mesh** removes the middleman entirely. Peers discover each other automatically using mDNS, connect directly over libp2p, and exchange messages and voice clips without touching the internet. The moment two devices join the same Wi-Fi, they can talk.

Built in Go. Runs in a browser. Zero configuration required.

---

## Features

-  **Text messaging** — sub-second delivery between peers on the same LAN
-  **Voice notes** — record and send audio clips directly from the browser (no plugin needed)
-  **Automatic peer discovery** — mDNS finds peers in 1–2 seconds, no IP addresses to exchange
-  **Encrypted transport** — libp2p upgrades connections to TLS 1.3 automatically
-  **Browser-based UI** — any modern browser works; nothing to install on the client side
-  **Zero internet dependency** — works on an isolated LAN with no external connectivity whatsoever

---

## How it works

Each node runs a single Go binary that handles three things simultaneously:

1. **mDNS agent** — announces presence and discovers other nodes on the LAN
2. **libp2p host** — manages direct peer connections, stream multiplexing, and transport encryption
3. **WebSocket server** — bridges the browser frontend to the P2P backend

When you send a message, the Go backend iterates over all known peers and opens a short-lived libp2p stream to each one — a full-mesh application-layer broadcast with no relay involved.

Voice notes follow a record → encode (Opus/WebM) → Base64 → JSON → libp2p stream → playback pipeline. It's a voice-note model rather than real-time streaming, which keeps the implementation simple and reliable.

```
Browser ──WebSocket──► Go Backend ──libp2p stream──► Remote Go Backend ──WebSocket──► Remote Browser
                           │                                   ▲
                           └──────── mDNS discovery ──────────┘
```

---

## Architecture

| Component | Role |
|---|---|
| Go / libp2p 0.32.0 | P2P host, stream multiplexing, TLS 1.3 transport |
| mDNS (port 5353) | Zero-config peer discovery on the local network |
| Gorilla WebSocket 1.5.1 | Browser ↔ Go backend bridge |
| MediaRecorder API | In-browser audio capture, Opus codec encoding |
| HTML/CSS/JS frontend | Served by the Go binary, runs in any modern browser |

**Network topology:** Full mesh. Every node connects directly to every other node. No hub, no relay, no DHT needed (mDNS is LAN-only by design).

**Ports used:**

| Port | Protocol | Purpose |
|---|---|---|
| `4001` | TCP/UDP | libp2p swarm (P2P data streams) |
| `5353` | UDP | mDNS peer discovery |
| `8080` | TCP | WebSocket server (browser UI) |

---

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- A modern browser (Chrome 110+ or Firefox 110+) for voice note support
- Two or more devices on the same local Wi-Fi network

---

## Getting started

```bash
# Clone the repository
git clone https://github.com/yashogale30/cssproject
cd cssproject

# Install dependencies
go mod tidy

# Run the node
go run main.go
```

Then open `http://localhost:8080` in your browser.

Repeat on any other device connected to the same Wi-Fi. Peers will appear automatically within a second or two — no configuration needed.

---

## Usage

1. Start the binary on each device
2. Open `http://localhost:8080` in a browser on each device
3. Watch peers appear in the peer list automatically
4. Type a message and hit Send, or click the mic button to record a voice note
5. Messages and audio appear on all connected peers instantly

---

## Testing

Tested on two laptops (macOS, ARM) on the same indoor Wi-Fi network.

| Test | Result |
|---|---|
| mDNS peer discovery | < 1–2 seconds consistently, across multiple restarts |
| Text message latency | < 1 second, no dropped messages observed |
| Voice note delivery | < 1 second for short clips; longer clips scale linearly with size |
| Re-discovery after restart | Peer re-appears within 1–2 seconds |

> **Note:** Testing was limited to two nodes due to hardware availability. Multi-node behavior (3+ peers) has not been empirically validated, though the architecture supports it.

---

## Limitations

- **No message persistence** — history lives in memory only; a page refresh clears it
- **LAN-only** — mDNS does not cross NAT boundaries or reach the public internet by design
- **No offline delivery** — messages sent while a peer is offline are lost
- **No end-to-end encryption** — transport is encrypted (TLS 1.3), but application-layer E2E is not yet implemented
- **No file transfer** — audio and text only for now (file transfer could be added as a new libp2p protocol stream)
- **Voice notes, not live calls** — audio is record-then-send, not real-time streaming

---

## Hardware performance guide

The application layer is rarely the bottleneck. Your router sets the ceiling:

| Hardware | Approx. concurrent users | Range |
|---|---|---|
| ESP32 / ESP8266 | 5–7 | 10–50 m |
| TP-Link N300 (802.11n) | 10–30 | 30–60 m |
| Jio AX6000 (Wi-Fi 6) | 80–150+ | 80–120 m |
| Mesh system | 150–300+ | 100–300 m |

---

## Tech stack

- **Backend:** Go, [libp2p](https://libp2p.io/), [Gorilla WebSocket](https://github.com/gorilla/websocket)
- **Frontend:** Vanilla HTML/CSS/JS, MediaRecorder API, WebSocket API
- **Discovery:** Multicast DNS (RFC 6762)
- **Audio codec:** Opus (via browser MediaRecorder, WebM container)
- **Transport security:** TLS 1.3 (automatic via libp2p)

---

## Acknowledgements

Built as part of the Communication in Service to Society (CSS) course. Thanks to the open-source communities behind libp2p, Go, and the browser platform APIs that made this practical to build.

---

## License

MIT
