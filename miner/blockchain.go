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
	BlockChainMap map[string] interface{}   // string is the hash of the block, interface is the block
	Heads []string                          // list of the hash of blocks on the longest chains
	initialized bool                        // indicates if the block chain has been initialized
	genesisHash   string
}

type Block interface {
	String() string							// function to get the string representation of the block which will be used to find the nonce of the block
	GetStringWithoutNonce() string          // function to get the string representation of the block without the nonce
	GetHash() string                        // function to get the hash base on the string representation of the block
	GetPrevHash() string                    // function to get the previous hash of the block
	// getCoinsRequirementInBlock: creates a map of key value pair with key being the miner Id, and value being the total
	//                             number of coins required for the miner to complete all the operations in the block
	GetCoinsRequirementInBlock() map[string]uint32
}

// Represents one block in the block chain
type OpBlock struct {
	Index      uint32
	PrevHash   string
	Sig        Signature
	Operations []interface{} // List of operations
	Nonce      uint32
}

type NoOpBlock struct {
	Index    uint32    // Index of the block
	PrevHash string    // MD5 hash of the previous bloc
	Sig      Signature // The signiture of the miner
	Nonce    uint32    // A 32 bit string as the nonce
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
	String() string                // Gets the string representation of the operation
	GetHash() string               // Gets the hash of the operation
	GetCost() uint32               // Gets the cost of the operation
	GetId() OpIdentity             // Gets the operation identity of the operation
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
	T    time.Time     // Time stamp of the operation
}

type Cryptopuzzle struct {
	Hash string // block hash without nonce
	N    int    // PoW difficulty: number of zeroes expected at end of md5
}

// isBlockValid: Checks if a block is valid base on below criterias
// 1. Check that the nonce for the block is valid: PoW is correct and has the right difficulty.
// 2. Check that the previous block hash points to a legal, previously generated, block.
// 3. Check the Index of the block is 1+ the index of the previous block
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
// minerId: the Id of the miner where we are trying to determine whether has sufficient coins
// coins: the minimum number of coins the miner should have
// returns: true if minerId has enough coins to carry out the operations in the block
func (bc *BlockChain) hasEnoughCoins(minerId string, coins uint32) bool {

	// TODO: Stub
	return false
}

// hasConflictingOperations: Check if there are conflicting operations along a particular chain
// b: the block we are trying to incorporate into the block chain. We will follow the PrevHash until we hit the genesis block
// returns: true if there are conflicting operations in block b along the block chain.
func (bc *BlockChain) hasConflictingOperations(b interface{}) bool {

	// TODO: Stub
	return false
}

// AddBlockToBlockChain: Adds the block to the block chain after validating the block
//                       1. Adds the block to BlockChainMap
//                       2. Remove the parents of the block from Heads list in the BlockChain object
//                       3. Add this block to the Heads list
func (bc *BlockChain) addBlockToBlockChain(b *Block){
	// TODO: Stub
}

// getChainBeforeBlock: Gets the longest block chain preceded the block b
// b: The block we are trying to find the chain prior to
// returns a list of hash of ordered blocks from the block chain that preceded block b
// Note: this function can only get called if b is validated to be a valid block in the chain
func(bc *BlockChain) getChainBeforeBlock(b interface{}) []string {

	prevHash := b.(Block).GetPrevHash()
	chain := []string{prevHash}

	for prevHash != bc.genesisHash {
		chain = append([]string{prevHash}, chain...)
		prevHash = bc.BlockChainMap[prevHash].(Block).GetPrevHash()
	}

	return chain
}

func (ob *OpBlock) GetStringWithoutNonce() string {

	str := string(ob.Index) + ob.PrevHash + ob.Sig.String()

	for _, op := range ob.Operations {
		str = str + op.(Operation).String()
	}

	return str
}

func (ob *OpBlock) String() string {

	return ob.GetStringWithoutNonce() + string(ob.Nonce)
}

func (nob *NoOpBlock) GetStringWithoutNonce() string {
	return string(nob.Index) + nob.PrevHash + nob.Sig.String()
}

func (nob *NoOpBlock) String() string {

	return nob.GetStringWithoutNonce() + string(nob.Nonce)
}

func (ob *OpBlock) GetHash() string {
	return GetMd5Hash(ob.String())
}

func (nob *NoOpBlock) GetHash() string {
	return GetMd5Hash(nob.String())
}

func (ob OpBlock) GetPrevHash() string {
	return ob.PrevHash
}

func (nob NoOpBlock) GetPrevHash() string {
	return nob.PrevHash
}

func (ob *OpBlock) GetCoinsRequirementInBlock() map[string]uint32 {
	coinReq := make(map[string]uint32)

	for _, op := range ob.Operations {
		minerId := op.(Operation).GetId().MinerId
		cost := op.(Operation).GetCost()
		if val, ok := coinReq[minerId]; ok {
			coinReq[minerId] = val + cost
		} else {
			coinReq[minerId] = cost
		}
	}

	return coinReq
}

func (nob *NoOpBlock) GetCoinsRequirementInBlock() map[string]uint32 {
	return make(map[string]uint32)
}

func (cf *CreateFile) String() string {
	return cf.OpId.String() + cf.FileName + string(cf.Cost)
}

func (ar *AppendRecord) String() string {
	return ar.OpId.String() + string(ar.Rec[:]) + string(ar.Cost) + ar.T.String()
}

func (cf *CreateFile) GetHash() string {
	return GetMd5Hash(cf.String())
}

func (ar *AppendRecord) GetHash() string {
	return GetMd5Hash(ar.String())
}

func (cf *CreateFile) GetCost() uint32 {
	return cf.Cost
}

func (ar *AppendRecord) GetCost() uint32 {
	return ar.Cost
}

func (cf *CreateFile) GetId() OpIdentity {
	return cf.OpId
}

func (ar *AppendRecord) GetId() OpIdentity {
	return ar.OpId
}

func(opid *OpIdentity) String() string {
	return opid.ClientId + opid.MinerId
}

func(sig *Signature) String() string {
	return sig.Id + string(sig.Coins)
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
