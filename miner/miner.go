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
	"time"
)

var TCP_PROTO = "tcp"
const MSG_SIZE = 4                          // 4 bytes in size

type Miner struct {
	config Config                           // Configuration of the miner
	pendingOps []interface{}                // A list of Operations, Operations could be CreateFile or AppendRecord
	bc *BlockChain                          // Placeholder for the block chain
	doNotForward map[string] bool           // A map that keeps track of the message that shouldn'T be forwarded, the string
	                                        // is the Hash of the message, the bool is a place holder that does not store any meaningful information
	peerChanOut map[string] chan *Message   // Map of peer ipport to channels that contains messages to be forwarded to the peers
	peerChanOutSig map[string] chan int     // Map of peer ipport to channels to send kill signals to incoming connection handler processes
	peerChanInSig map[string] chan int      // Map of peer ipport to channels to send kill signals to outgoing connection handler processes
	peerChanIn chan *Message                // A chanel to receive message from peers, different handler will be called depending on different messages
	lock *sync.Mutex                        // The mutex to synchronize the data structure access of processes
}

// Represents the configuration of the miner, the configuration will be loaded from a JSON file
type Config struct {
	PeerMinersAddrs        []string // A list of TCP IP:port strings that are the set of peer miners that this miner should connect to.
	IncomingMinersAddr     string   // The TcpIp:port where this miner can receive connections from other miners.
	OutgoingMinersIP       string   // The local IP that the miner should use to connect to peer miners
	IncomingClientsAddr    string   // The TcpIp:port where this miner can receive connections from rfs clients.
	MinerID                string   // The ID of this miner
	MinedCoinsPerOpBlock   uint32   // The number of record Coins mined for an op block
	MinedCoinsPerNoOpBlock uint32   // The number of record Coins mined for a no-op block
	NumCoinsPerFileCreate uint32    // The number of record Coins charged for creating a file
	GenesisBlockHash string         // The genesis (first) block MD5 Hash for this blockchain
	PowPerOpBlock uint32            // The op block difficulty (proof of work setting: number of zeroes)
	PowPerNoOpBlock uint32          // The no-op block difficulty (proof of work setting: number of zeroes)
	ConfirmsPerFileCreate uint32    // The number of confirmations for a create file operation (the number of blocks that must follow the block containing a create file operation along longest chain before the CreateFile call can return successfully)
	ConfirmsPerFileAppend uint32    // The number of confirmations for an append operation (the number of blocks that must follow the block containing an append operation along longest chain before the AppendRec call can return successfully)
	GenOpBlockTimeout uint32        // Time in milliseconds, the minimum time between op block mining
}

// Container used to send data over the network, Content is the serialized Operations, Block or BlockChain to be sent
// across the network
type Message struct {
	MsgType uint16 // 0 is the entire block chain, 1 is op block, 2 is NoOpBlock, 3 is CreateFile operation,
											// 4 is AppendRecord operation, 5 is request to send the entire block chain
	T       time.Time // Time stamp set by the client or the miner represents the client in order to differentiate messages
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


	return &Miner{config, make([]interface{}, 0), &BlockChain{make(map[string] interface{}), make([]string, 0), false, ""}, make(map[string] bool),
	make(map[string] chan *Message), make(map[string] chan int), make(map[string] chan int), make(chan *Message), &sync.Mutex{}}, nil
}

// AcceptPeerConnections: Accepts peer connections on the IncomingMinersAddr specified in the JSON configuration file
func (m *Miner) acceptPeerConnections() {
	for {
		localTcpAddr, err := net.ResolveTCPAddr(TCP_PROTO, m.config.IncomingMinersAddr)
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
		m.peerChanOut[peerIpPort] = msgChan

		killSigIn := make(chan int)
		m.peerChanInSig[peerIpPort] = killSigIn

		killSigOut := make(chan int)
		m.peerChanOutSig[peerIpPort] = killSigOut

		m.lock.Unlock()

		go m.handlePeerConnectionIn(peerIpPort, tcpConn, killSigIn)
		go m.handlePeerConnectionOut(peerIpPort, tcpConn, msgChan, killSigOut)
	}
}

// StartPeerConnections: starts connections to peers specified in the JSON configuration file
func(m *Miner) startPeerConnections() {
	tcpLocalAddr, err := net.ResolveTCPAddr(TCP_PROTO, m.config.OutgoingMinersIP)
	if err != nil {
		panic("Unable to resolve local TCP address")
	}

	for _, peerIpPort := range m.config.PeerMinersAddrs {
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
		m.peerChanOut[peerIpPort] = msgChan

		killSigIn := make(chan int)
		m.peerChanInSig[peerIpPort] = killSigIn

		killSigOut := make(chan int)
		m.peerChanOutSig[peerIpPort] = killSigOut

		m.lock.Unlock()

		go m.handlePeerConnectionIn(peerIpPort, tcpConn, killSigIn)
		go m.handlePeerConnectionOut(peerIpPort, tcpConn, msgChan, killSigOut)

		reqForBlockChain := Message{5, time.Now().UTC(),make([]byte,0)}

		// Send message to peers to request for the entire block chain
		msgChan <- &reqForBlockChain
	}
}

// NotifyPeers: sends the message to all peers outgoing channels so the out handler will forward the message to the
//              corresponding peers
// msg: the message to be sent
func(m *Miner) notifyPeers(msg *Message) {
	for _, v := range m.peerChanOut {
		v <- msg
	}
}

// HandlePeerConnectionOut: waiting to receive message from the channel and forward the messages to peers
//                          when the connection with the peer is closed, this process will be killed.
//                          this process will handle the connection closure so HandlePeerConnectionIn doesn'T have to
//                          worry about it.
// conn: the TCP connection with one peer
// c: the channel of message that this process waits to receive message from
// Sig: the channel this process receives the kill signal
func (m *Miner) handlePeerConnectionOut(peerIpPort string, conn *net.TCPConn, msgChan chan *Message, sig chan int) {
	for {
		select {
		case <-sig:
			fmt.Printf("Kill signal for HandlePeerConnectionOut received for %s, stopping thread.\n", peerIpPort)
			break
		case msg := <- msgChan:
			forward := true
			// Check if we should forward the message
			msgBytes, err := msg.getBytesFromMsg()

			if err != nil {
				fmt.Println("Get message Hash failed.")
				continue
			}

			hash := GetMd5Hash(string(msgBytes))

			m.lock.Lock()
			if _, ok := m.doNotForward[hash]; ok {
				forward = false
			}

			if forward {
				m.doNotForward[hash] = true
				sendMsgToTcp(conn, msg)
			}
			m.lock.Unlock()
		default:
			// stop for one second
			// time.Sleep(1 * time.Second)
		}
	}

	fmt.Printf("Thread successfully stopped for HandlePeerConnectionOut for %s, performing clean up.\n", peerIpPort)
	conn.Close()
	close(msgChan)
	close(sig)

	m.lock.Lock()
	delete(m.peerChanOutSig, peerIpPort)
	delete(m.peerChanOut, peerIpPort)
	m.lock.Unlock()
}

// HandlePeerConnectionIn: waiting to receive message from the tcp connection, write this message to the out going channel
//                         of all peers and handle the message depending on message types
// conn: the TCP connection with one peer
// Sig: the channel this process receives the kill signal
func (m *Miner) handlePeerConnectionIn(peerIpPort string, conn *net.TCPConn, sig chan int) {

	for {
		select {
		case <-sig:
			fmt.Printf("Kill signal for HandlePeerConnectionIn received for %s, stopping thread.\n", peerIpPort)
			break
		default:
		}

		msg, err := readMsgFromTcp(conn)

		if err != nil {
			fmt.Println("Message read failed.")
			continue
		}

		switch msg.MsgType {
		// TODO: wait for a while, if no block chain received, need to start mining no op blocks on its own
		// TODO: this to do is just a reminder, the actual implementation should be else where
		case 0:
			if !m.bc.Initialized {
				var bc BlockChain
				err := json.Unmarshal(msg.Content, &bc)
				if err != nil {
					fmt.Println("Decoding block chain failed")
					continue
				}

				m.bc = &bc
			}

		case 1:
			var ob OpBlock
			err := json.Unmarshal(msg.Content, &ob)
			if err != nil {
				fmt.Println("Decoding op block failed")
				continue
			}

			// If not exist in the DoNotForward map, then forward to peers
			msgBytes, err := msg.getBytesFromMsg()

			if err != nil {
				fmt.Println("Get message Hash failed.")
				continue
			}

			hash := GetMd5Hash(string(msgBytes))

			if _, ok := m.doNotForward[hash]; !ok {
				m.lock.Lock()
				m.doNotForward[hash] = true
				m.lock.Unlock()
				m.notifyPeers(msg)
			}

			// TODO: then incorporate the block into the block chain

		case 2:
			var nob NoOpBlock
			err := json.Unmarshal(msg.Content, &nob)
			if err != nil {
				fmt.Println("Decoding no op block failed")
				continue
			}

			msgBytes, err := msg.getBytesFromMsg()

			if err != nil {
				fmt.Println("Get message Hash failed.")
				continue
			}

			hash := GetMd5Hash(string(msgBytes))

			// If not exist in the DoNotForward map, then forward to peers
			if _, ok := m.doNotForward[hash]; !ok {
				m.lock.Lock()
				m.doNotForward[hash] = true
				m.lock.Unlock()
				m.notifyPeers(msg)
			}

			// TODO: then incorporate the block into the block chain

		case 3:
			var cf CreateFile
			err := json.Unmarshal(msg.Content, &cf)
			if err != nil {
				fmt.Println("Decoding no create file operation failed")
				continue
			}

			msgBytes, err := msg.getBytesFromMsg()

			if err != nil {
				fmt.Println("Get message Hash failed.")
				continue
			}

			hash := GetMd5Hash(string(msgBytes))

			// If not exist in the DoNotForward map, then forward to peers
			if _, ok := m.doNotForward[hash]; !ok {
				m.lock.Lock()
				m.doNotForward[hash] = true
				// TODO: add this operation to list if the message has not been seen before
				m.pendingOps = append(m.pendingOps, cf)
				m.lock.Unlock()
				m.notifyPeers(msg)
			}

		case 4:
			var ar AppendRecord
			err := json.Unmarshal(msg.Content, &ar)
			if err != nil {
				fmt.Println("Decoding no create file operation failed")
				continue
			}

			msgBytes, err := msg.getBytesFromMsg()

			if err != nil {
				fmt.Println("Get message Hash failed.")
				continue
			}

			hash := GetMd5Hash(string(msgBytes))

			// If not exist in the DoNotForward map, then forward to peers
			if _, ok := m.doNotForward[hash]; !ok {
				m.lock.Lock()
				m.doNotForward[hash] = true
				// TODO: add this operation to list if the message has not been seen before
				m.pendingOps = append(m.pendingOps, ar)
				m.lock.Unlock()
				m.notifyPeers(msg)
			}

		case 5:
			bcBytes, err := json.Marshal(*m.bc)

			if err != nil {
				fmt.Println("Encoding of block chain failed")
				continue
			}
			msg := Message{0, time.Now().UTC(), bcBytes}
			sendMsgToTcp(conn, &msg)
		}
	}

	fmt.Printf("Thread successfully stopped for HandlePeerConnectionIn for %s, performing clean up.\n", peerIpPort)

	close(sig)

	m.lock.Lock()
	delete(m.peerChanInSig, peerIpPort)
	m.lock.Unlock()
}

func readMsgFromTcp(conn *net.TCPConn) (*Message, error){
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

func sendMsgToTcp(conn *net.TCPConn, msg *Message) error {

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

// createOpBlock: Create the next Op block in the block chain
// prevBlock: is the previous block of the current block that the miner is trying to create
// c: is the channel of message where if a block is received and is the same level as the block the miner is currently
//    creating, the miner needs to drop the block it's trying to create and incorporate the block into the block chain.
//    after that, the miner continues to mine blocks, either no-op blocks or op blocks.
func (m *Miner) createOpBlock(prevBlock *Block, c chan *Message) (*OpBlock, error) {
	//STUB
	return &OpBlock{}, nil
}

// createNoOpBlock: Create the next No op block in the block chain
// prevBlock: is the previous block of the current block that the miner is trying to create
// c: is the channel of message where if a block is received and is the same level as the block the miner is currently
//    creating, the miner needs to drop the block it's trying to create and incorporate the block into the block chain.
//    after that, the miner continues to mine blocks, either no-op blocks or op blocks.
func (m *Miner) createNoOpBlock(prevBlock *Block, c chan *Message) (*NoOpBlock, error) {
	return &NoOpBlock{}, nil
}


func (m *Miner) StartProcess() {
	m.startPeerConnections()
	go m.acceptPeerConnections()

	// Wait for 30 seconds to receive the block chain from peers
	time.Sleep(30 * time.Second)

	// TODO: if block chain is not Initialized, initialize it and start generating blocks

	// TODO: else, just start generating blocks
}

func (ms *Message) getBytesFromMsg() ([]byte, error) {
	msgTypeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(msgTypeBytes, uint16(ms.MsgType))

	timeStampBytes, err := ms.T.MarshalBinary()

	if err != nil {
		fmt.Println("Time marshaling failed in message")
		return nil, err
	}
	res := append(msgTypeBytes, timeStampBytes...)
	res = append(res, ms.Content...)
	return res, nil
}

