// TODO: All below codes are subject to change base on the block and message design, be careful

package miner

import (
	"cpsc416-p1/rfslib"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
)

var TCP_PROTO = "tcp"

type Miner struct {
	Config Config                           // Configuration of the miner
	PendingOps []Operation                  // A list of pending operations
	BC *BlockChain                          // Placeholder for the block chain
	DoNotForward map[string] uint32         // A map that keeps track of the message that shouldn't be forwarded, the string
	                                        // is the hash of the message, the uint32 is the count of number of times the miner
	                                        // expects to receive the same message back (this number will be initialized as the number of peers this miner has),
	                                        // every time it receives the message, this number will be decremented by one,
	                                        // and when it reaches 0, it can be deleted from the map
	PeerChan map[string] chan *Message
}

// Represents the configuration of the miner, the configuration will be loaded from a JSON file
type Config struct {
	Peers []string                          // A list of TCP IP:port strings that are the set of peer miners that this miner should connect to.
	LocalTcpIpPort string                   // The TcpIp:port where this miner can receive connections from other miners.
	ClientTcpIpPort string                  // The TcpIp:port where this miner can receive connections from rfs clients.
	Id string                               // The ID of this miner
	NumCoinsOpBlock uint32                  // The number of record coins mined for an op block
	NumCoinsNopBlock uint32                 // The number of record coins mined for a no-op block
	NumCoinsPerFile uint32                  // The number of record coins charged for creating a file
	GenesisHash string                      // The genesis (first) block MD5 hash for this blockchain
	OpDifficulty uint32                     // The op block difficulty (proof of work setting: number of zeroes)
	NopDifficulty uint32                    // The no-op block difficulty (proof of work setting: number of zeroes)
	NumConfFc uint32                        // The number of confirmations for a create file operation (the number of blocks that must follow the block containing a create file operation along longest chain before the CreateFile call can return successfully)
	NumConfApp uint32                       // The number of confirmations for an append operation (the number of blocks that must follow the block containing an append operation along longest chain before the AppendRec call can return successfully)
}

type BlockChain struct {
	Roots []Node
	// Miner needs to maintain this map and will have to create this map on its own when first joining the network
	BlockChainMap map[string] *Block        // string is the hash of the block
	Heads []*Block                          // list of heads on the longest chains
}

type Node struct {
	data Block
	Children []Node
}

// Represents one block in the block chain
type Block struct {
	Index uint32                            // Index of the block
	PrevHash string                         // MD5 hash of the previous block
	Operations []Operation                  // List of operations
	Sig Signature							// The signiture of the miner
	nonce string                            // A 32 bit string as the nonce
}

// Represents the signature of the block
type Signature struct {
	Id string                               // The miner Id that this credit goes to
	Coins uint32                            // Number of coins that's awarded
}

// The operation interface, CreateFile and AppendRecord are both operations and must implement the Operation interface
// so we can pass them around as type of Operation
type Operation interface {
	isSame(other interface{}) bool          // Compares if two operations are the same, they are considered to be the same if the operation Id's are the same between two operations
}

type OpIdentity struct {
	ClientId string                         // Id of the client that submitted the operation
	MinerId string                          // Id of the miner who submitted the operation on be half of the client
	Type uint8                              // 3 is create file, 4 is append record
}

// Two CreateFile operations are the same if the file names are the same
type CreateFile struct {
	OpId Operation
	FileName string                         // Name of the file that we are creating
	Cost uint32                             // Cost of creating the file
}

// Two AppendRecord operations are the same if the ClientId are the same and the Rec are the same
type AppendRecord struct {
	OpId Operation
	Rec rfslib.Record                       // 512 byte record
	Cost uint32                             // Cost of appending the record (always 1 coin)
}

// Container used to send data over the network, Content is the serialized Operations, Block or BlockChain to be sent
// across the network
// TODO: to send the entire block chain, we will have to send it over block by block
type Message struct {
	Type uint8                              // 0 is the entire block chain, 1 is one block, 2 is an operation, 5 is the end of message used to mark the transmission of the entire blockchain is over
	                                        // 6 is request to send the entire block chain
	Content []byte
}

func InitializeMiner(pathToJson string) (*Miner, error){
	jsonFile, err := os.Open(pathToJson)

	if err != nil {
		fmt.Println("Miner: Error opening the configuration file, please try again.")
		return nil, err
	}

	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)

	if err != nil {
		fmt.Println("Miner: Error reading the configuration file, please try again")
		return nil, err
	}

	var config Config

	err = json.Unmarshal(bytes, &config)

	if err != nil {
		fmt.Println("Miner: Error parsing the configuration file, please try again")
		return nil, err
	}


	return &Miner{config, make([]Operation,0), new(BlockChain), make(map[string] uint32 ), make(map[string] chan *Message)}, nil
}

// Create the next block in the block chain
// prevBlock: is the previous block of the current block that the miner is trying to create
// c: is the channel of message where if a block is received and is the same level as the block the miner is currently
//    creating, the miner needs to drop the block it's trying to create and incorporate the block into the block chain.
//    after that, the miner continues to mine blocks, either no-op blocks or op blocks.
// TODO: might need to include a chanel for interupt
func (m *Miner) createBlock(prevBlock *Block, c chan *Message) (Block, error) {
	//STUB
	return Block{}, nil
}


// isBlockValid: Checks if a block is valid base on below criterias
// 1. Check that the nonce for the block is valid: PoW is correct and has the right difficulty.
// 2. Check that the previous block hash points to a legal, previously generated, block.
// 3. Check that each operation in the block is associated with a miner ID that has enough record
//	  coins to pay for the operation (i.e., the number of record coins associated with the minerID
//	  must have sufficient balance to 'pay' for the operation).
// 4. Check that each operation does not violate RFS semantics (e.g., a record is not mutated or
//    inserted into the middled of an rfs file).
// 5. We need to check there is no conflicting operations along the chain in the block. i.e. there are no two create
//    operations with the same file name.
func (m *Miner) isBlockValid(b *Block) bool {

	//STUB
	return false
}


// getCoinsRequirementInBlock: creates a map of key value pair with key being the miner Id, and value being the total
//                             number of coins required for the miner to complete all the operations in the block
// b: the block which we are trying to calculate the coin requirements for each miner.
func getCoinsRequirementInBlock(b *Block) map[string] uint32 {


	return make(map[string] uint32)
}


// hasEnoughCoins: Determines if a miner has enough coins
// minerId: the Id of the miner where we are trying to determine whether has sufficient coins
// coins: the minimum number of coins the miner should have
// returns: true if minerId has enough coins to carry out the operations in the block
func (m *Miner) hasEnoughCoins(minerId string, coins uint32) bool {

	//STUB
	return false
}


// hasConflictingOperations: Check if there are conflicting operations along a particular chain
// b: the block we are trying to incorporate into the block chain. We will follow the PrevHash until we hit the genesis block
// returns: true if there are conflicting operations in block b along the block chain.
func (m *Miner) hasConflictingOperations(b *Block) bool {

	//STUB
	return false
}


// AddBlockToBlockChain: Adds the block to the block chain after validating the block
//                       1. Adds the block to the tree structure
//                       2. Adds the block to BlockChainMap
//                       3. Remove the parents of the block from Heads list in the BlockChain object
//                       4. Add this block to the Heads list
func (m *Miner) AddBlockToBlockChain(b *Block) {

}

// getNextBlockSize: Helper function to generate a number between the min and max value, used to determine the number of
// records to obtain from PendingOps
// min: is the minimum block size
// max: is the maximum block size
func getNextBlockSize(min int8, max int8) int8{
	//STUB
	return 0
}


// Helper function to get the string representation of the block which will be used to find the nonce of the block
// b: is the pointer to the block that we need to create the string representation of.
func getStringFromBlock(b *Block) string {

	//STUB
	return ""
}

//TODO: The client and other peers may be using the same TCP connection to connect, need a way to differentiate.
//TODO: For now we assume we only have peers connect via this connection.
// AcceptPeerConnections: Accepts peer connections on the IpPort specified in the JSON configuration file
func (m *Miner) AcceptPeerConnections() error {
	localTcpAddr, err := net.ResolveTCPAddr(TCP_PROTO, m.Config.LocalTcpIpPort)
	if err != nil {
		fmt.Println("Listener creation failed, please try again.")
		return err
	}
	listener, err := net.ListenTCP(TCP_PROTO, localTcpAddr)

	conn, err := listener.AcceptTCP()
	if err != nil {
		fmt.Println("TCP connection failed.")
		return err
	}

	go m.HandlePeerConnectionIn(conn)

	return nil
}

// StartPeerConnections: starts connections to peers specified in the JSON configuration file
func(m *Miner) StartPeerConnections() {
	tcpLocalAddr, err := net.ResolveTCPAddr(TCP_PROTO, m.Config.LocalTcpIpPort)
	if err != nil {
		panic("Unable to resolve local TCP address")
	}

	for _, ipPort := range m.Config.Peers {
		tcpOutAddr, err := net.ResolveTCPAddr(TCP_PROTO, ipPort)
		if err != nil {
			fmt.Println("Unable to resolve peer IpPort:", ipPort)
			continue
		}

		tcpConn, err := net.DialTCP(TCP_PROTO,tcpOutAddr, tcpLocalAddr)

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


// SendBlockChain: Sends the blockchain block by block via the TCP connection
// conn: the TCP connection to send the block chain to
// return: error if there are problems sending any of the blocks via the TCP connection
//         nil if the entire block chain is successfully transmitted
func (m *Miner) SendBlockChain(conn *net.TCPConn) error {

	return errors.New("")
}


// CreateBlockChain: Creates the entire block chain object after receiving all the blocks from it's peer
// blocks: is the map of blocks with key being the hash of the block and value being a pointer to the block
func (m *Miner) CreateBlockChain(blocks map[string] *Block) {

}



