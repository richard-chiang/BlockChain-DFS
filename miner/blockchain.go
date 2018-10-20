package miner

import (
	"cpsc416-p1/rfslib"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

var mutexBC = &sync.Mutex{}                 // the mutex for the block chain data structure

type BlockChain struct {
	// Miner needs to maintain this map and will have to create this map on its own when first joining the network
	blockChainMap map[string] interface{} // string is the Hash of the block, interface is the block
	heads         []string                // list of the Hash of blocks on the longest chains
	initialized   bool                    // indicates if the block chain has been initialized
	genesisHash   string
}

type Block interface {
	String() string                // function to get the string representation of the block which will be used to find the nonce of the block
	getStringWithoutNonce() string // function to get the string representation of the block without the nonce
	getHash() string               // function to get the Hash base on the string representation of the block
	getPrevHash() string           // function to get the previous Hash of the block
	// getCoinsRequirementInBlock: creates a map of key value pair with key being the miner id, and value being the total
	//                             number of coins required for the miner to complete all the operations in the block
	getCoinsRequirementInBlock() map[string]uint32
}

// Represents one block in the block chain
type OpBlock struct {
	index      uint32
	prevHash   string
	sig        Signature
	operations []interface{} // List of operations
	nonce      uint32
}

type NoOpBlock struct {
	index    uint32    // index of the block
	prevHash string    // MD5 Hash of the previous bloc
	sig      Signature // The signiture of the miner
	nonce    uint32    // A 32 bit string as the nonce
}

// Represents the signature of the block
type Signature struct {
	id    string // The miner id that this credit goes to
	coins uint32 // Number of coins that's awarded
}

// The operation interface, CreateFile and AppendRecord are both operations and must implement the Operation interface
// so we can pass them around as type of Operation
type Operation interface {
	isSame(other interface{}) bool // Compares if two operations are the same, they are considered to be the same if the operation id's are the same between two operations
	String() string                // Gets the string representation of the operation
	getHash() string               // Gets the Hash of the operation
	getCost() uint32               // Gets the cost of the operation
	getId() OpIdentity             // Gets the operation identity of the operation
}

type OpIdentity struct {
	clientId string // id of the client that submitted the operation
	minerId  string // id of the miner who submitted the operation on be half of the client
}

// Two CreateFile operations are the same if the file names are the same
type CreateFile struct {
	opId     OpIdentity
	fileName string // Name of the file that we are creating
	cost     uint32 // cost of creating the file
}

// Two AppendRecord operations are the same if the clientId are the same and the rec are the same
type AppendRecord struct {
	opId OpIdentity
	rec  rfslib.Record // 512 byte record
	cost uint32        // cost of appending the record (always 1 coin)
	t    time.Time     // Time stamp of the operation
}

type Cryptopuzzle struct {
	Hash string // block Hash without nonce
	N    int    // PoW difficulty: number of zeroes expected at end of md5
}

// isBlockValid: Checks if a block is valid base on below criterias
// 1. Check that the nonce for the block is valid: PoW is correct and has the right difficulty.
// 2. Check that the previous block Hash points to a legal, previously generated, block.
// 3. Check the index of the block is 1+ the index of the previous block
// 4. Check that each operation in the block is associated with a miner ID that has enough record
//	  coins to pay for the operation (i.e., the number of record coins associated with the minerID
//	  must have sufficient balance to 'pay' for the operation).
// 5. Check that each operation does not violate RFS semantics (e.g., a record is not mutated or
//    inserted into the middled of an rfs file).
// 6. We need to check there is no conflicting operations along the chain in the block. i.e. there are no two create
//    operations with the same file name.
func (bc *BlockChain) isBlockValid(b *Block) bool {

	// TODO: Stub
	return false
}

// hasEnoughCoins: Determines if a miner has enough coins
// minerId: the id of the miner where we are trying to determine whether has sufficient coins
// coins: the minimum number of coins the miner should have
// returns: true if minerId has enough coins to carry out the operations in the block
func (bc *BlockChain) hasEnoughCoins(minerId string, coins uint32) bool {

	// TODO: Stub
	return false
}

func (bc *BlockChain) getCoinBalanceInChain(chain []string) map[string] uint32 {
	balance := make(map[string] uint32)
	// TODO: Stub
	return balance
}

// hasConflictingOperations: Check if there are conflicting operations along a particular chain
// b: the block we are trying to incorporate into the block chain. We will follow the prevHash until we hit the genesis block
// returns: true if there are conflicting operations in block b along the block chain.
func (bc *BlockChain) hasConflictingOperations(b interface{}) bool {

	// TODO: Stub
	return false
}

// AddBlockToBlockChain: Adds the block to the block chain after validating the block
//                       1. Adds the block to blockChainMap
//                       2. Remove the parents of the block from heads list in the BlockChain object
//                       3. Add this block to the heads list
func (bc *BlockChain) addBlockToBlockChain(b *Block){
	// TODO: Stub
}

// getChainBeforeBlock: Gets the longest block chain preceded the block b
// b: The block we are trying to find the chain prior to
// returns a list of Hash of ordered blocks from the block chain that preceded block b
// Note: this function can only get called if b is validated to be a valid block in the chain
func(bc *BlockChain) getChainBeforeBlock(b interface{}) []string {

	prevHash := b.(Block).getPrevHash()
	chain := []string{prevHash}

	for prevHash != bc.genesisHash {
		chain = append([]string{prevHash}, chain...)
		prevHash = bc.blockChainMap[prevHash].(Block).getPrevHash()
	}

	return chain
}

func (ob *OpBlock) getStringWithoutNonce() string {

	str := string(ob.index) + ob.prevHash + ob.sig.String()

	for _, op := range ob.operations {
		str = str + op.(Operation).String()
	}

	return str
}

func (ob *OpBlock) String() string {

	return ob.getStringWithoutNonce() + string(ob.nonce)
}

func (nob *NoOpBlock) getStringWithoutNonce() string {
	return string(nob.index) + nob.prevHash + nob.sig.String()
}

func (nob *NoOpBlock) String() string {

	return nob.getStringWithoutNonce() + string(nob.nonce)
}

func (ob *OpBlock) getHash() string {
	return GetMd5Hash(ob.String())
}

func (nob *NoOpBlock) getHash() string {
	return GetMd5Hash(nob.String())
}

func (ob OpBlock) getPrevHash() string {
	return ob.prevHash
}

func (nob NoOpBlock) getPrevHash() string {
	return nob.prevHash
}

func (ob *OpBlock) getCoinsRequirementInBlock() map[string]uint32 {
	coinReq := make(map[string]uint32)

	for _, op := range ob.operations {
		minerId := op.(Operation).getId().minerId
		cost := op.(Operation).getCost()
		if val, ok := coinReq[minerId]; ok {
			coinReq[minerId] = val + cost
		} else {
			coinReq[minerId] = cost
		}
	}

	return coinReq
}

func (nob *NoOpBlock) getCoinsRequirementInBlock() map[string]uint32 {
	return make(map[string]uint32)
}

func (cf *CreateFile) String() string {
	return cf.opId.String() + cf.fileName + string(cf.cost)
}

func (ar *AppendRecord) String() string {
	return ar.opId.String() + string(ar.rec[:]) + string(ar.cost) + ar.t.String()
}

func (cf *CreateFile) getHash() string {
	return GetMd5Hash(cf.String())
}

func (ar *AppendRecord) getHash() string {
	return GetMd5Hash(ar.String())
}

func (cf *CreateFile) getCost() uint32 {
	return cf.cost
}

func (ar *AppendRecord) getCost() uint32 {
	return ar.cost
}

func (cf *CreateFile) getId() OpIdentity {
	return cf.opId
}

func (ar *AppendRecord) getId() OpIdentity {
	return ar.opId
}

func(opid *OpIdentity) String() string {
	return opid.clientId + opid.minerId
}

func(sig *Signature) String() string {
	return sig.id + string(sig.coins)
}

func GetMd5Hash(input string) string {
	h := md5.New()
	h.Write([]byte(input))
	res := hex.EncodeToString(h.Sum(nil))
	return res
}

func CalcSecret(problem Cryptopuzzle) uint32 {
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
	return GetMd5Hash(hash + string(nonce))
}
