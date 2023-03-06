package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	peerstore "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	multiaddr "github.com/multiformats/go-multiaddr"
)

func main() {
	node, err := libp2p.New(
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/0"),
		libp2p.Ping(false),
	)

	if err != nil {
		panic(err)
	}

	fmt.Println("Id: ", node.ID(), "\nAddresses: ", node.Addrs())

	// configure our own ping protocol
	pingService := &ping.PingService{Host: node}
	node.SetStreamHandler(ping.ID, func(s network.Stream) {
		defer s.Close()
		conn := s.Conn()

		pingService.PingHandler(s)
		fmt.Printf("\nGot pings from Peer(Id: %s, Address: %s)", conn.RemotePeer(), conn.RemoteMultiaddr())
	})

	// print the node's PeerInfo in multiaddr format
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, _ := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	fmt.Println("libp2p node address:", addrs[0])
	fmt.Println("-------------------------------")

	if len(os.Args) > 1 {
		// Get Multiaddr from args
		addr, err := multiaddr.NewMultiaddr(os.Args[1])
		if err != nil {
			panic(err)
		}

		// Get peer addrInfo from multiaddr
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			panic(err)
		}

		// Connect to the peer
		if err := node.Connect(context.Background(), *peer); err != nil {
			panic(err)
		}

		fmt.Println("sending 5 ping messages to", addr)
		ch := pingService.Ping(context.Background(), peer.ID)
		for i := 0; i < 5; i++ {
			res := <-ch
			fmt.Println("Pinged!! ", "RTT:", res.RTT)
		}
	} else {
		// wait for a SIGINT or SIGTERM signal
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		fmt.Println("\nReceived signal, shutting down...")
	}

	// Shutdown the node
	if err := node.Close(); err != nil {
		panic(err)
	}
}
