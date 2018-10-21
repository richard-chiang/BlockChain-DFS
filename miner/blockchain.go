package miner

import (
	"cpsc416-p1/rfslib"
	"crypto/md5"
	"encoding/hex"
	"math"
	"strings"
	"sync"
	"time"
)

var mutexBC = &sync.Mutex{}                 // the mutex for the block chain data structure

type BlockChain struct {
	// Miner needs to maintain this map and will have to create this map on its own when first joining the network
	BlockChainMap map[string] interface{} // string is the Hash of the block, interface is the block
	Heads         []string                // list of the Hash of blocks on the longest chains
	Initialized   bool                    // indicates if the block chain has been Initialized
	GenesisHash   string
}

type Block interface {
	String() string                // function to get the string representation of the block which will be used to find the Nonce of the block
	getStringWithoutNonce() string // function to get the string representation of the block without the Nonce
	getHash() string               // function to get the Hash base on the string representation of the block
	getPrevHash() string           // function to get the previous Hash of the block
	getOps() []interface{}         // gets the list of Operations in the block
	getSig() *Signature            // gets the signature of the block
	// getCoinsRequirementInBlock: creates a map of key value pair with key being the miner Id, and value being the total
	//                             number of Coins required for the miner to complete all the Operations in the block
	getCoinsRequirementInBlock() map[string]int
}

// Represents one block in the block chain
type OpBlock struct {
	Index      uint32
	PrevHash   string
	Sig        Signature
	Operations []interface{} // List of Operations
	Nonce      uint32
}

type NoOpBlock struct {
	Index    uint32    // Index of the block
	PrevHash string    // MD5 Hash of the previous bloc
	Sig      Signature // The signiture of the miner
	Nonce    uint32    // A 32 bit string as the Nonce
}

// Represents the signature of the block
type Signature struct {
	Id    string // The miner Id that this credit goes to
	Coins int    // Number of Coins that's awarded
}

// The operation interface, CreateFile and AppendRecord are both Operations and must implement the Operation interface
// so we can pass them around as type of Operation
type Operation interface {
	//isSame(other interface{}) bool // Compares if two Operations are the same, they are considered to be the same if the operation Id's are the same between two Operations
	String() string                // Gets the string representation of the operation
	getHash() string               // Gets the Hash of the operation
	getCost() int                  // Gets the Cost of the operation
	getId() OpIdentity             // Gets the operation identity of the operation
}

type OpIdentity struct {
	ClientId string // Id of the client that submitted the operation
	MinerId  string // Id of the miner who submitted the operation on be half of the client
}

// Two CreateFile Operations are the same if the file names are the same
type CreateFile struct {
	OpId     OpIdentity
	FileName string // Name of the file that we are creating
	Cost     int    // Cost of creating the file
}

// Two AppendRecord Operations are the same if the ClientId are the same and the Rec are the same
type AppendRecord struct {
	OpId OpIdentity
	Rec  rfslib.Record // 512 byte record
	Cost int           // Cost of appending the record (always 1 coin)
	T    time.Time     // Time stamp of the operation
}

type Cryptopuzzle struct {
	Hash string // block Hash without Nonce
	N    int    // PoW difficulty: number of zeroes expected at end of md5
}

// isBlockValid: Checks if a block is valid base on below criterias
// 1. Check that the Nonce for the block is valid: PoW is correct and has the right difficulty.
// 2. Check that the previous block Hash points to a legal, previously generated, block.
// 3. Check the Index of the block is 1+ the Index of the previous block
// 4. Check that each operation in the block is associated with a miner ID that has enough record
//	  Coins to pay for the operation (i.e., the number of record Coins associated with the minerID
//	  must have sufficient balance to 'pay' for the operation).
// 5. Check that each operation does not violate RFS semantics (e.g., a record is not mutated or
//    inserted into the middled of an rfs file).
// 6. We need to check there is no conflicting Operations along the chain in the block. i.e. there are no two create
//    Operations with the same file name.
func (bc *BlockChain) isBlockValid(b Block) bool {

	// TODO: Stub
	chain := bc.getChainBeforeBlock(b)
	req := b.getCoinsRequirementInBlock()

	if bc.hasEnoughCoins(chain, req, b) && !bc.hasConflictingOperations(b) {
		return true
	}

	return false
}

// hasEnoughCoins: Determines if a miner has enough Coins
// req: the map of key value pair of miner Id to required Coins
// b: the pointer to the block that we are trying to add to the block chain
// returns: true if MinerId has enough Coins to carry out the Operations in the block
func (bc *BlockChain) hasEnoughCoins(chain []string, req map[string] int, b Block) bool {

	balance := bc.getCoinBalanceInChain(chain, req, b)

	for minerId, reqCoins := range req {
		if bal, ok := balance[minerId]; !ok || bal < reqCoins{
			return false
		}
	}

	return true
}

// getCoinBalanceInChain: gets the balance along the chain for each miner in the req map
// chain: a list of string that represents the chain we are interested in
// req: the map of key value pair of miner Id to required Coins
// b: the pointer to the block that we are trying to add to the block chain
// returns: a map of miner Id to coin balance
func (bc *BlockChain) getCoinBalanceInChain(chain []string, req map[string] int, b Block) map[string] int {
	balance := make(map[string] int)

	// Initialize balance with same miner Id with req and 0 as starting balance
	for minerId := range req {
		balance[minerId] = 0
	}

	for _, str := range chain {
		block, _ := bc.BlockChainMap[str]
		ops := block.(Block).getOps()
		sig := block.(Block).getSig()

		// Add to balance
		if _, ok := req[sig.Id]; ok {
			balance[sig.Id] += sig.Coins
		}

		// Subtract from balance
		for _, op := range ops {
			id := op.(Operation).getId()
			cost := op.(Operation).getCost()

			if _, ok := req[id.MinerId]; ok {
				balance[id.MinerId] -= cost
			}
		}
	}

	if coinsCurrentBlock, ok := req[b.getSig().Id]; ok {
		balance[b.getSig().Id] += coinsCurrentBlock
	}

	return balance
}

// hasConflictingOperations: Check if there are conflicting Operations along a particular chain
// b: the block we are trying to incorporate into the block chain. We will follow the PrevHash until we hit the genesis block
// returns: true if there are conflicting Operations in block b along the block chain.
func (bc *BlockChain) hasConflictingOperations(b Block) bool {

	// TODO: Stub
	return false
}

// AddBlockToBlockChain: Adds the block to the block chain after validating the block
//                       1. Adds the block to BlockChainMap
//                       2. Remove the parents of the block from Heads list in the BlockChain object
//                       3. Add this block to the Heads list
func (bc *BlockChain) addBlockToBlockChain(b Block){
	// TODO: Stub
}

// getChainBeforeBlock: Gets the longest block chain preceded the block b
// b: The block we are trying to find the chain prior to
// returns a list of Hash of ordered blocks from the block chain that preceded block b
// Note: this function can only get called if b is validated to be a valid block in the chain
func(bc *BlockChain) getChainBeforeBlock(b interface{}) []string {

	prevHash := b.(Block).getPrevHash()
	chain := make([]string, 0)

	for prevHash != bc.GenesisHash {
		chain = append([]string{prevHash}, chain...)
		prevHash = bc.BlockChainMap[prevHash].(Block).getPrevHash()
	}

	return chain
}

func (ob OpBlock) getStringWithoutNonce() string {

	str := string(ob.Index) + ob.PrevHash + ob.Sig.String()

	for _, op := range ob.Operations {
		str = str + op.(Operation).String()
	}

	return str
}

func (ob OpBlock) String() string {

	return ob.getStringWithoutNonce() + string(ob.Nonce)
}

func (nob NoOpBlock) getStringWithoutNonce() string {
	return string(nob.Index) + nob.PrevHash + nob.Sig.String()
}

func (nob NoOpBlock) String() string {

	return nob.getStringWithoutNonce() + string(nob.Nonce)
}

func (ob OpBlock) getHash() string {
	return GetMd5Hash(ob.String())
}

func (nob NoOpBlock) getHash() string {
	return GetMd5Hash(nob.String())
}

func (ob OpBlock) getPrevHash() string {
	return ob.PrevHash
}

func (nob NoOpBlock) getPrevHash() string {
	return nob.PrevHash
}

func (ob OpBlock) getCoinsRequirementInBlock() map[string]int {
	coinReq := make(map[string]int)

	for _, op := range ob.Operations {
		minerId := op.(Operation).getId().MinerId
		cost := op.(Operation).getCost()
		if val, ok := coinReq[minerId]; ok {
			coinReq[minerId] = val + cost
		} else {
			coinReq[minerId] = cost
		}
	}

	return coinReq
}

func (nob NoOpBlock) getCoinsRequirementInBlock() map[string]int {
	return make(map[string]int)
}

func (ob OpBlock) getOps() []interface{} {
	return ob.Operations
}

func (nob NoOpBlock) getOps() []interface{} {
	return make([]interface{},0)
}

func (ob OpBlock) getSig() *Signature {
	return &ob.Sig
}

func (nob NoOpBlock) getSig() *Signature {
	return &nob.Sig
}

func (cf CreateFile) String() string {
	return cf.OpId.String() + cf.FileName + string(cf.Cost)
}

func (ar AppendRecord) String() string {
	return ar.OpId.String() + string(ar.Rec[:]) + string(ar.Cost) + ar.T.String()
}

func (cf CreateFile) getHash() string {
	return GetMd5Hash(cf.String())
}

func (ar AppendRecord) getHash() string {
	return GetMd5Hash(ar.String())
}

func (cf CreateFile) getCost() int {
	return cf.Cost
}

func (ar AppendRecord) getCost() int {
	return ar.Cost
}

func (cf CreateFile) getId() OpIdentity {
	return cf.OpId
}

func (ar AppendRecord) getId() OpIdentity {
	return ar.OpId
}

func(opid OpIdentity) String() string {
	return opid.ClientId + opid.MinerId
}

func(sig Signature) String() string {
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
