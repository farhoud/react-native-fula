package fula

import (
	"context"
	"fmt"
	"log"
	"os"

	libp2p "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/peerstore"
	mplex "github.com/libp2p/go-libp2p-mplex"
	noise "github.com/libp2p/go-libp2p-noise"
	"fulamobile/libfula/fula/protocol"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
)

type IFula interface {
	Connect()
	Send()
	ReceiveMeta()
}

type Fula struct {
	node  host.Host
	peers []peer.ID
}

func NewFula() *Fula {

	node, err := create()
	if err != nil {
		panic(err)
	}

	f := &Fula{node: node}

	return f
	
}

func (f *Fula) Connect(boxAddr string) {
	peerAddr, _ := peer.AddrInfoFromString(boxAddr)
	node := f.node
	f.peers = append(f.peers, peerAddr.ID)
	node.Peerstore().AddAddrs(peerAddr.ID, peerAddr.Addrs, peerstore.PermanentAddrTTL)
}

func (f *Fula) Send(filePath string) string{
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	stream, err := f.node.NewStream(context.Background(), f.peers[0], protocol.Protocol)
	if err != nil {
		panic(err)
	}
	fmt.Println("are u runing")
	res := protocol.SendFile(file, stream)
	fmt.Println(res)

	return res
}

func (f *Fula) ReceiveMeta(fileId string) string{
	stream, err := f.node.NewStream(context.Background(), f.peers[0], protocol.Protocol)
	if err != nil {
		panic(err)
	}
	res := protocol.ReceiveMeta(stream, fileId)
	fmt.Println("res", res)
	return res.String()
}

func create() (host.Host, error) {
	// Now, normally you do not just want a simple host, you want
	// that is fully configured to best support your p2p application.
	// Let's create a second host setting some more options.

	// Set your own keypair
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519, // Select your key type. Ed25519 are nice short
		-1,             // Select key length when possible (i.e. RSA).
	)
	if err != nil {
		panic(err)
	}
	con, err := connmgr.NewConnManager(1, 1)
	if err != nil {
		panic(err)
	}
	node, err := libp2p.New(
		// Use the keypair we generated
		libp2p.Identity(priv),
		// Multiple listen addresses
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/9000",      // regular tcp connections
			"/ip4/0.0.0.0/udp/9000/quic", // a UDP endpoint for the QUIC transport
		),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(con),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
		// Let this host use relays and advertise itself on relays if
		// it finds it is behind NAT. Use libp2p.Relay(options...) to
		// enable active relays and more.
		// libp2p.EnableAutoRelay(),
		libp2p.DisableRelay(),
		// If you want to help other peers to figure out if they are behind
		// NATs, you can launch the server-side of AutoNAT too (AutoRelay
		// already runs the client)
		//
		// This service is highly rate-limited and should not cause any
		// performance issues.
		libp2p.EnableNATService(),
	)

	log.Printf("Hello World, my second hosts ID is %s\n", node.ID())
	protocol.RegisterFileProtocol(node)
	return node, err
}
