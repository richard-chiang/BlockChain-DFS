// TODO: All below codes are subject to change base on the block and message design, be careful

package miner

import (
	"cpsc416-p1"
	"encoding/json"
	"fmt"
	"net"
)

var TCP_NETWORK = "tcp"
type Node interface {
	StartPeerConnections()
}

type Miner struct {
	Peers []string                          // A list of TCP IP:port strings that are the set of peer miners that this miner should connect to.
	TcpIpPort string                        // The TCP IP:port where this miner can receive connections from rfs clients.
	Id string                               // The ID of this miner
	NumCoinsOpBlock uint32                  // The number of record coins mined for an op block
	NumCoinsNopBlock uint32                 // The number of record coins mined for a no-op block
	NumCoinsPerFile uint32                  // The number of record coins charged for creating a file
	GenesisHash string                      // The genesis (first) block MD5 hash for this blockchain
	OpDifficulty uint32                     // The op block difficulty (proof of work setting: number of zeroes)
	NopDifficulty uint32                    // The no-op block difficulty (proof of work setting: number of zeroes)
	NumConfFc uint32                        // The number of confirmations for a create file operation (the number of blocks that must follow the block containing a create file operation along longest chain before the CreateFile call can return successfully)
	NumConfApp uint32                       // The number of confirmations for an append operation (the number of blocks that must follow the block containing an append operation along longest chain before the AppendRec call can return successfully)
	BlockChain map[string] string           // Placeholder for the block chain
	PeerChan map[string] chan *Message
}


type Block struct {
	PrevHash string
	Records []rfslib.Record
	nonce uint32
}

type Message struct {
	Type uint8              // TODO: assume 0 is flood for now
	Content []byte
}


//TODO: The client and other peers may be using the same TCP connection to connect, need a way to differentiate.
//TODO: For now we assume we only have peers connect via this connection.
func (m *Miner) Accept() error {
	localTcpAddr, err := net.ResolveTCPAddr(TCP_NETWORK, m.TcpIpPort)
	if err != nil {
		fmt.Println("Listener creation failed, please try again.")
		return err
	}
	listener, err := net.ListenTCP(TCP_NETWORK, localTcpAddr)

	conn, err := listener.AcceptTCP()
	if err != nil {
		fmt.Println("TCP connection failed.")
		return err
	}

	go m.HandlePeerConnectionIn(conn)

	return nil
}

func(m *Miner) StartPeerConnections() {
	tcpLocalAddr, err := net.ResolveTCPAddr(TCP_NETWORK, m.TcpIpPort)
	if err != nil {
		panic("Unable to resolve local TCP address")
	}

	for _, ipPort := range m.Peers {
		tcpOutAddr, err := net.ResolveTCPAddr(TCP_NETWORK, ipPort)
		if err != nil {
			fmt.Println("Unable to resolve peer IpPort:", ipPort)
			continue
		}

		tcpConn, err := net.DialTCP(TCP_NETWORK,tcpOutAddr, tcpLocalAddr)

		if err != nil {
			fmt.Println("Failed to establish connection with peer:", ipPort)
			continue
		}

		c := make(chan *Message)
		m.PeerChan[ipPort] = c

		go HandlePeerConnectionOut(tcpConn, c)
	}
}

func(m *Miner) NotifyPeers(msg *Message) {
	for _, v := range m.PeerChan {
		v <- msg
	}
}

// TODO: Example code, need changes in the future.
func HandlePeerConnectionOut(conn *net.TCPConn, c chan *Message) {
	defer conn.Close()

	for {
		select {
			case msg := <- c:
				b, e := json.Marshal(*msg)
				if e != nil {
					fmt.Println("Message encoding failed")
					continue
				}
				conn.Write(b)
		}
	}
}

func (m *Miner) HandlePeerConnectionIn(conn *net.TCPConn) {

	for {
		// TODO: We need to design the message structure first before we can finalize the size, now just hard coding
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("read timeout:", err)
				// time out
			} else {
				fmt.Println("read error:", err)
				// some error else, do something else, for example create new conn
			}
		}

		var msg Message

		json.Unmarshal(buf[:n], msg)

		if msg.Type == 0 {
			m.NotifyPeers(&msg)
		}
	}
}


