package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/gordonklaus/portaudio"
	"github.com/libp2p/go-libp2p/core/network"
)

const protocolID = "/p2p-voice/1.0.0"

func main() {

	ctx := context.Background()

	host, err := libp2p.New()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Peer ID:", host.ID())

	for _, addr := range host.Addrs() {
		fmt.Println("Listening on:", addr)
	}

	host.SetStreamHandler(protocolID, handleStream)

	err = setupMDNS(ctx, host)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("mDNS discovery started")

	reader := bufio.NewReader(os.Stdin)

	for {

		fmt.Println("\nPress ENTER to record voice message")
		reader.ReadString('\n')

		audio := recordAudio()

		for _, peer := range host.Network().Peers() {

			stream, err := host.NewStream(ctx, peer, protocolID)
			if err != nil {
				continue
			}

			stream.Write(audio)
			stream.Close()

			fmt.Println("Voice message sent to", peer)
		}
	}
}

func handleStream(s network.Stream) {

	fmt.Println("Receiving voice message")

	data, err := io.ReadAll(s)
	if err != nil {
		return
	}

	playAudio(data)
}

func recordAudio() []byte {
	portaudio.Initialize()
	defer portaudio.Terminate()

	sampleRate := 44100
	seconds := 10
	chunkSize := sampleRate 

	
	in := make([]int16, sampleRate*seconds)

	chunk := make([]int16, chunkSize)


	stream, err := portaudio.OpenDefaultStream(1, 0, float64(sampleRate), len(chunk), &chunk)
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	stream.Start()


	fmt.Print("🔴 RECORDING NOW: ")


	for i := 0; i < seconds; i++ {
		err := stream.Read()
		if err != nil {
			log.Println("Error reading audio:", err)
		}

	
		copy(in[i*chunkSize:], chunk)


		fmt.Print(seconds-i, "... ")
	}

	stream.Stop()
	fmt.Println("\n Done!")


	buf := make([]byte, len(in)*2)
	for i, v := range in {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(v))
	}

	return buf
}

func playAudio(data []byte) {

	portaudio.Initialize()
	defer portaudio.Terminate()

	out := make([]int16, len(data)/2)

	for i := range out {
		out[i] = int16(binary.LittleEndian.Uint16(data[i*2:]))
	}

	stream, err := portaudio.OpenDefaultStream(0, 1, 44100, 512, &out)
	if err != nil {
		log.Println(err)
		return
	}

	stream.Start()
	stream.Write()
	stream.Stop()
}