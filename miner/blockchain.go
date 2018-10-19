package miner

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"../rfslib"
)

const (
	appendCost uint32 = 2
	createCost uint32 = 1
)

var mutexBC = &sync.Mutex{} // the mutex for the block chain data structure

type BlockChain struct {
	// Miner needs to maintain this map and will have to create this map on its own when first joining the network
	BlockChainMap map[string]interface{} // string is the hash of the block, interface is the block
	Heads         []string               // list of the hash of blocks on the longest chains
	initialized   bool                   // indicates if the block chain has been initialized
	genesisHash   string
}

type Block interface {
	// function to get the string representation of the block which will be used to find the nonce of the block
	getStringFromBlock() string

	// getCoinsRequirementInBlock: creates a map of key value pair with key being the miner Id, and value being the total
	//                             number of coins required for the miner to complete all the operations in the block
	getCoinsRequirementInBlock() map[string]uint32
	isSame(other interface{}) bool
	getPreviousHash() string
}

// Represents one block in the block chain
type OpBlock struct {
	Index      uint32
	PrevHash   string
	Sig        Signature
	Operations []interface{} // List of operations
	nonce      uint32
}

type NoOpBlock struct {
	Index    uint32    // Index of the block
	PrevHash string    // MD5 hash of the previous bloc
	Sig      Signature // The signiture of the miner
	nonce    uint32    // A 32 bit string as the nonce
}

// Represents the signature of the block
type Signature struct {
	Id    string // The miner Id that this credit goes to
	Coins uint32 // Number of coins that's awarded
}

// The operation interface, CreateFile and AppendRecord are both operations and must implement the Operation interface
// so we can pass them around as type of Operation
type Operation interface {
	isSame(other interface{}) bool // Compares if two operations are the same, they are considered to be the same if the operation Id's are the same between two operations
	getStringFromOp() string       // Gets the string re
	getOpID() OpIdentity
}

type OpIdentity struct {
	ClientId string // Id of the client that submitted the operation
	MinerId  string // Id of the miner who submitted the operation on be half of the client
}

// Two CreateFile operations are the same if the file names are the same
type CreateFile struct {
	OpId     OpIdentity
	FileName string // Name of the file that we are creating
	Cost     uint32 // Cost of creating the file
}

// Two AppendRecord operations are the same if the ClientId are the same and the Rec are the same
type AppendRecord struct {
	OpId OpIdentity
	Rec  rfslib.Record // 512 byte record
	Cost uint32        // Cost of appending the record (always 1 coin)
	t    time.Time     // Time stamp of the operation
}

type cryptopuzzle struct {
	Hash string // block hash without nonce
	N    int    // PoW difficulty: number of zeroes expected at end of md5
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
func (bc *BlockChain) isBlockValid(b *Block) bool {

	//STUB
	return false
}

// hasEnoughCoins: Determines if a miner has enough coins
// minerId: the Id of the miner where we are trying to determine whether has sufficient coins
// coins: the minimum number of coins the miner should have
// returns: true if minerId has enough coins to carry out the operations in the block
func (bc *BlockChain) hasEnoughCoins(minerId string, coins uint32) bool {

	//STUB
	return false
}

// hasConflictingOperations: Check if there are conflicting operations along a particular chain
// b: the block we are trying to incorporate into the block chain. We will follow the PrevHash until we hit the genesis block
// returns: true if there are conflicting operations in block b along the block chain.
func (bc *BlockChain) hasConflictingOperations(b *interface{}) bool {

	//STUB
	return false
}

// AddBlockToBlockChain: Adds the block to the block chain after validating the block
//                       1. Adds the block to BlockChainMap
//                       2. Remove the parents of the block from Heads list in the BlockChain object
//                       3. Add this block to the Heads list
func (bc *BlockChain) addBlockToBlockChain(b *Block) {
	//stub
}

// getLongestChain: Gets the longest block chain preceded the block b
// b: the block we are trying to find the chain prior to
// returns a list of hash of ordered blocks from the block chain that preceded block b
func (bc *BlockChain) getLongestChain(b interface{}) []interface{} {
	block := b.(Block)
	blockchain := bc.BlockChainMap
	startingHash := bc.genesisHash
	heads := bc.Heads

	for _, headHash := range heads {
		bIsInThisHash := false
		hashChain := bc.generateChain(headHash)

		for _, blockHash := range hashChain {
			if blockchain[blockHash].(Block).isSame() {
				bIsInThisHash = true
				break
			}
		}
	}

	// stub

	return make([]interface{}, 0)
}

func (ob *OpBlock) getStringFromBlock() string {
	// Edwin todo
	return ""
}

func (nob *NoOpBlock) getStringFromBlock() string {
	// Edwin todo
	return ""
}

func (ob *OpBlock) getCoinsRequirementInBlock() map[string]uint32 {
	page := make(map[string]uint32)

	for _, op := range ob.Operations {
		minerID := op.(Operation).getOpID().MinerId
		switch v := op.(type) {
		case CreateFile:
			page[minerID] += createCost
		case AppendRecord:
			page[minerID] += appendCost
		}
	}

	return page
}

func (nob *NoOpBlock) getCoinsRequirementInBlock() map[string]uint32 {
	return make(map[string]uint32)
}

func (cf CreateFile) getOpID() OpIdentity {
	return cf.OpId
}

func (ar AppendRecord) getOpID() OpIdentity {
	return ar.OpId
}

func (cf CreateFile) getStringFromOp() string {
	op := cf.OpId
	return op.ClientId + op.MinerId + cf.FileName + fmt.Sprint(cf.Cost)
}

func (ar AppendRecord) getStringFromOp() string {
	op := ar.OpId
	opString := op.ClientId + op.MinerId
	recString := string(ar.Rec[:])
	costString := fmt.Sprint(ar.Cost)
	timeString := ar.t.String()
	return opString + recString + costString + timeString
}

func (ob OpBlock) isSame(other interface{}) bool {
	otherBlock := other.(Block)

	myHash := ob.getStringFromBlock()
	otherHash := otherBlock.getStringFromBlock()
	return myHash == otherHash
}

func (nob NoOpBlock) isSame(other interface{}) bool {
	otherBlock := other.(Block)

	myHash := nob.getStringFromBlock()
	otherHash := otherBlock.getStringFromBlock()
	return myHash == otherHash
}

func (ob OpBlock) getPreviousHash() string {
	return ob.PrevHash
}

func (nob NoOpBlock) getPreviousHash() string {
	return nob.PrevHash
}

// given the hash of a head
// return the chain from the head to genesis hash
func (bc BlockChain) generateChain(head string) []string {
	first := bc.BlockChainMap[head].(Block)
	second := first.getPreviousHash()

	var collector []string
	for blockHash := head; blockHash != bc.genesisHash; blockHash = bc.BlockChainMap[blockHash].(Block).getPreviousHash() {
		collector = append(collector, blockHash)
	}

	collector = append(collector, bc.genesisHash)
	return collector
}

func getMd5Hash(input string) string {
	h := md5.New()
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func calcSecret(problem cryptopuzzle) uint32 {
	result := ""
	var nonce uint32

	for nonce = 0; nonce < math.MaxUint32; nonce++ {
		result = computeNonceSecretHash(problem.Hash, nonce)

		if validNonce(problem.N, result) {
			fmt.Println(result)
			return nonce
		}
	}
	return nonce
}

func validNonce(N int, Hash string) bool {
	zeros := strings.Repeat("0", N)
	isValid := strings.HasSuffix(Hash, zeros)
	return isValid
}

func computeNonceSecretHash(hash string, nonce uint32) string {
	return getMd5Hash(hash + string(nonce))
}
