// TODO: All below codes are subject to change base on the block and message design, be careful

package miner

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sync"
)

var TCP_PROTO = "tcp"
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const MSG_SIZE = 4                          // 4 bytes in size

type Miner struct {
	Config Config                           // Configuration of the miner
	PendingOps []Operation                  // A list of pending operations
	BC *BlockChain                          // Placeholder for the block chain
	DoNotForward map[string] uint32         // A map that keeps track of the message that shouldn't be forwarded, the string
	                                        // is the hash of the message, the uint32 is the count of number of times the miner
	                                        // expects to receive the same message back (this number will be initialized as the number of peers this miner has),
	                                        // every time it receives the message, this number will be decremented by one,
	                                        // and when it reaches 0, it can be deleted from the map
	PeerChanOut map[string] chan *Message   // Map of peer ipport to channels that contains messages to be forwarded to the peers
	PeerChanOutSig map[string] chan int     // Map of peer ipport to channels to send kill signals to incoming connection handler processes
	PeerChanInSig map[string] chan int      // Map of peer ipport to channels to send kill signals to outgoing connection handler processes
	PeerChanIn chan *Message                // A chanel to receive message from peers, different handler will be called depending on different messages
	lock *sync.Mutex                        // The mutex to synchronize the data structure access of processes
}

// Represents the configuration of the miner, the configuration will be loaded from a JSON file
type Config struct {
	PeerMinersAddrs []string                // A list of TCP IP:port strings that are the set of peer miners that this miner should connect to.
	IncomingMinersAddr string               // The TcpIp:port where this miner can receive connections from other miners.
	OutgoingMinersIP string                 // The local IP that the miner should use to connect to peer miners
	IncomingClientsAddr string              // The TcpIp:port where this miner can receive connections from rfs clients.
	MinerID string                          // The ID of this miner
	MinedCoinsPerOpBlock uint32             // The number of record coins mined for an op block
	MinedCoinsPerNoOpBlock uint32           // The number of record coins mined for a no-op block
	NumCoinsPerFileCreate uint32            // The number of record coins charged for creating a file
	GenesisBlockHash string                 // The genesis (first) block MD5 hash for this blockchain
	PowPerOpBlock uint32                    // The op block difficulty (proof of work setting: number of zeroes)
	PowPerNoOpBlock uint32                  // The no-op block difficulty (proof of work setting: number of zeroes)
	ConfirmsPerFileCreate uint32            // The number of confirmations for a create file operation (the number of blocks that must follow the block containing a create file operation along longest chain before the CreateFile call can return successfully)
	ConfirmsPerFileAppend uint32            // The number of confirmations for an append operation (the number of blocks that must follow the block containing an append operation along longest chain before the AppendRec call can return successfully)
	GenOpBlockTimeout uint32                // Time in milliseconds, the minimum time between op block mining
}

// Container used to send data over the network, Content is the serialized Operations, Block or BlockChain to be sent
// across the network
type Message struct {
	Type uint8                              // 0 is the entire block chain, 1 is op block, 2 is NoOpBlock, 3 is CreateFile operation,
											// 4 is AppendRecord operation, 5 is request to send the entire block chain
	Content []byte
}

func InitializeMiner(pathToJson string) (*Miner, error){
	jsonFile, err := os.Open(pathToJson)

	if err != nil {
		fmt.Println("Miner: Error opening the configuration file, please try again.")
		return nil, err
	}

	defer jsonFile.Close()

	configData, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Miner: Error reading the configuration file, please try again")
		return nil, err
	}

	var config Config

	err = json.Unmarshal(configData, &config)

	if err != nil {
		fmt.Println("Miner: Error parsing the configuration file, please try again")
		return nil, err
	}


	return &Miner{config, make([]Operation,0), &BlockChain{make(map[uint32] Block), make([]uint32, 0), false}, make(map[string] uint32 ),
	make(map[string] chan *Message), make(map[string] chan int), make(map[string] chan int), make(chan *Message), &sync.Mutex{}}, nil
}

// AcceptPeerConnections: Accepts peer connections on the IncomingMinersAddr specified in the JSON configuration file
func (m *Miner) AcceptPeerConnections() {
	for {
		localTcpAddr, err := net.ResolveTCPAddr(TCP_PROTO, m.Config.IncomingMinersAddr)
		if err != nil {
			fmt.Println("Listener creation failed, please try again.")
			return
		}

		listener, err := net.ListenTCP(TCP_PROTO, localTcpAddr)

		tcpConn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("TCP connection failed.")
			return
		}

		peerIpPort:= tcpConn.RemoteAddr().String()

		m.lock.Lock()

		msgChan := make(chan *Message)
		m.PeerChanOut[peerIpPort] = msgChan

		killSigIn := make(chan int)
		m.PeerChanInSig[peerIpPort] = killSigIn

		killSigOut := make(chan int)
		m.PeerChanOutSig[peerIpPort] = killSigOut

		m.lock.Unlock()

		go m.HandlePeerConnectionIn(peerIpPort, tcpConn, killSigIn)
		go m.HandlePeerConnectionOut(peerIpPort, tcpConn, msgChan, killSigOut)
	}
}

// StartPeerConnections: starts connections to peers specified in the JSON configuration file
func(m *Miner) StartPeerConnections() {
	tcpLocalAddr, err := net.ResolveTCPAddr(TCP_PROTO, m.Config.OutgoingMinersIP)
	if err != nil {
		panic("Unable to resolve local TCP address")
	}

	for _, peerIpPort := range m.Config.PeerMinersAddrs {
		tcpPeerAddr, err := net.ResolveTCPAddr(TCP_PROTO, peerIpPort)
		if err != nil {
			fmt.Println("Unable to resolve peer IpPort:", peerIpPort)
			continue
		}

		tcpConn, err := net.DialTCP(TCP_PROTO, tcpLocalAddr, tcpPeerAddr)

		if err != nil {
			fmt.Println("Failed to establish connection with peer:", peerIpPort)
			continue
		}

		m.lock.Lock()

		msgChan := make(chan *Message)
		m.PeerChanOut[peerIpPort] = msgChan

		killSigIn := make(chan int)
		m.PeerChanInSig[peerIpPort] = killSigIn

		killSigOut := make(chan int)
		m.PeerChanOutSig[peerIpPort] = killSigOut

		m.lock.Unlock()

		go m.HandlePeerConnectionIn(peerIpPort, tcpConn, killSigIn)
		go m.HandlePeerConnectionOut(peerIpPort, tcpConn, msgChan, killSigOut)

		reqForBlockChain := Message{5, make([]byte,0)}

		// Send message to peers to request for the entire block chain
		msgChan <- &reqForBlockChain
	}
}

// NotifyPeers: sends the message to all peers outgoing channels so the out handler will forward the message to the
//              corresponding peers
// msg: the message to be sent
func(m *Miner) NotifyPeers(msg *Message) {
	for _, v := range m.PeerChanOut {
		v <- msg
	}
}

// HandlePeerConnectionOut: waiting to receive message from the channel and forward the messages to peers
//                          when the connection with the peer is closed, this process will be killed.
//                          this process will handle the connection closure so HandlePeerConnectionIn doesn't have to
//                          worry about it.
// conn: the TCP connection with one peer
// c: the channel of message that this process waits to receive message from
// sig: the channel this process receives the kill signal
func (m *Miner) HandlePeerConnectionOut(peerIpPort string, conn *net.TCPConn, msgChan chan *Message, sig chan int) {
	for {
		select {
		case <-sig:
			fmt.Printf("Kill signal for HandlePeerConnectionOut received for %s, stopping thread.\n", peerIpPort)
			break
		case msg := <- msgChan:
			forward := true
			// Check if we should forward the message
			hash := getMd5Hash(string(msg.Type) + string(msg.Content))
			m.lock.Lock()
			if _, ok := m.DoNotForward[hash]; ok {
				//TODO: think about this, after one hop, the peer might expect 1 less flood of the same message
				m.DoNotForward[hash]--
				forward = false
			}
			m.lock.Unlock()

			if forward {
				SendMsgToTcp(conn, msg)
			}
		default:
			// stop for one second
			// time.Sleep(1 * time.Second)
		}
	}

	fmt.Printf("Thread successfully stopped for %s, performing clean up.\n", peerIpPort)
	conn.Close()
	close(msgChan)
	close(sig)

	m.lock.Lock()
	delete(m.PeerChanOutSig, peerIpPort)
	delete(m.PeerChanOut, peerIpPort)
	m.lock.Unlock()
}

// HandlePeerConnectionIn: waiting to receive message from the tcp connection, write this message to the out going channel
//                         of all peers and handle the message depending on message types
// conn: the TCP connection with one peer
// sig: the channel this process receives the kill signal
func (m *Miner) HandlePeerConnectionIn(peerIpPort string, conn *net.TCPConn, sig chan int) {

	for {
		msg, err := ReadMsgFromTcp(conn)

		if err != nil {
			fmt.Println("Message read failed.")
			continue
		}

		switch msg.Type {
		// TODO: wait for a while, if no block chain received, need to start mining no op blocks on its own
		case 0:
			if !m.BC.initialized {
				var bc BlockChain
				err := json.Unmarshal(msg.Content, &bc)
				if err != nil {
					fmt.Println("Decoding block chain failed")
					continue
				}

				m.BC = &bc
			}

		case 1:
			var ob OpBlock
			err := json.Unmarshal(msg.Content, &ob)
			if err != nil {
				fmt.Println("Decoding op block failed")
				continue
			}

			// If not exist in the DoNotForward map, then forward to peers
			if _, ok := m.DoNotForward[getMd5Hash(ob.getStringFromBlock())]; !ok {
				m.NotifyPeers(msg)
			}

			// TODO: then incorporate the block into the block chain

		case 2:
			var nob NoOpBlock
			err := json.Unmarshal(msg.Content, &nob)
			if err != nil {
				fmt.Println("Decoding no op block failed")
				continue
			}

			// If not exist in the DoNotForward map, then forward to peers
			if _, ok := m.DoNotForward[getMd5Hash(nob.getStringFromBlock())]; !ok {
				m.NotifyPeers(msg)
			}

			// TODO: then incorporate the block into the block chain

		case 3:
			var cf CreateFile
			err := json.Unmarshal(msg.Content, &cf)
			if err != nil {
				fmt.Println("Decoding no create file operation failed")
				continue
			}

			// If not exist in the DoNotForward map, then forward to peers
			if _, ok := m.DoNotForward[getMd5Hash(cf.getStringFromOp())]; !ok {
				m.NotifyPeers(msg)
			}

			// TODO: check if this op is already in PendingOps, if not, add it to the end of the list

		case 4:
			var ar AppendRecord
			err := json.Unmarshal(msg.Content, &ar)
			if err != nil {
				fmt.Println("Decoding no create file operation failed")
				continue
			}

			// If not exist in the DoNotForward map, then forward to peers
			if _, ok := m.DoNotForward[getMd5Hash(ar.getStringFromOp())]; !ok {
				m.NotifyPeers(msg)
			}

			// TODO: check if this op is already in PendingOps, if not, add it to the end of the list

		case 5:
			bcBytes, err := json.Marshal(*m.BC)

			if err != nil {
				fmt.Println("Encoding of block chain failed")
				continue
			}
			msg := Message{0, bcBytes}
			SendMsgToTcp(conn, &msg)
		}
	}
}

// SendBlockChain: Sends the entire block chain via the TCP connection
// conn: the TCP connection to send the block chain to
// return: error if there are problems sending any of the blocks via the TCP connection
//         nil if the entire block chain is successfully transmitted
func (m *Miner) SendBlockChain(conn *net.TCPConn) error {

	return errors.New("")
}

func ReadMsgFromTcp(conn *net.TCPConn) (*Message, error){
	var msg Message
	sizeBuf := make([]byte, MSG_SIZE)
	_, err := conn.Read(sizeBuf)

	if err != nil {
		return nil, err
	}

	mlen := binary.LittleEndian.Uint32(sizeBuf)

	if err != nil {
		return nil, err
	}

	msgBuff := make([]byte, mlen)

	sizeMsg, err := conn.Read(msgBuff)

	if uint32(sizeMsg) != mlen {
		return nil, errors.New("msg size wrong")
	}

	err = json.Unmarshal(msgBuff, &msg)

	if err != nil {
		return nil, err
	}

	return &msg, nil
}

func SendMsgToTcp(conn *net.TCPConn, msg *Message) error {

	msgBytes, err := json.Marshal(*msg)

	if err != nil {
		return err
	}

	b := make([]byte, MSG_SIZE)
	binary.LittleEndian.PutUint32(b, uint32(len(msgBytes)))

	conn.Write(b)

	conn.Write(msgBytes)

	return nil
}

