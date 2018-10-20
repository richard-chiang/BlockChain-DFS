package miner

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestMinerInitialization(t *testing.T) {
	jsonPath := "miner1.json"
	miner, err := InitializeMiner(jsonPath)

	if err != nil {
		fmt.Println(err)
		t.Errorf("error incured when initializing the miner.")
	}

	expectedPerrs := []string{"123.123.123.123:100", "234.234.234.234:200", "345,345,345.345:300"}

	for idx := 0; idx < len(expectedPerrs); idx++ {
		if expectedPerrs[idx] != miner.config.PeerMinersAddrs[idx] {
			t.Errorf("Field PeerMinersAddrs is not corrected parsed.")
		}
	}

	if miner.config.IncomingMinersAddr != "127.0.0.1:8081" {
		t.Errorf("Field IncomingMinersAddr is not corrected parsed.")
	}

	if miner.config.OutgoingMinersIP != "127.0.0.1:8082" {
		t.Errorf("Field OutgoingMinersIP is not corrected parsed.")
	}

	if miner.config.IncomingClientsAddr != "127.0.0.1:8080" {
		t.Errorf("Field IncomingClientsAddr is not corrected parsed.")
	}

	if miner.config.MinerID != "123456" {
		t.Errorf("Field MinerID is not correctly parsed.")
	}

	if miner.config.MinedCoinsPerOpBlock != 2 {
		t.Errorf("Field MinedCoinsPerOpBlock is not correctly parsed")
	}

	if miner.config.MinedCoinsPerNoOpBlock != 3 {
		t.Errorf("Field MinedCoinsPerNoOpBlock is not correctly parsed")
	}

	if miner.config.NumCoinsPerFileCreate != 1 {
		t.Errorf("Field NumCoinsPerFileCreate is not correctly parsed")
	}

	if miner.config.GenesisBlockHash != "qwerwerqwfsdfsadfsdf" {
		t.Errorf("Field GenesisBlockHash is not correctly parsed")
	}

	if miner.config.PowPerOpBlock != 6 {
		t.Error("Field PowPerOpBlock is not correctly parsed")
	}

	if miner.config.PowPerNoOpBlock != 5 {
		t.Error("Field PowPerNoOpBlock is not correctly parsed")
	}

	if miner.config.ConfirmsPerFileCreate != 5 {
		t.Error("Field ConfirmsPerFileCreate is not correctly parsed")
	}

	if miner.config.ConfirmsPerFileAppend != 6 {
		t.Error("Field ConfirmsPerFileAppend is not correctly parsed")
	}

	if miner.config.GenOpBlockTimeout != 5 {
		t.Error("Field GenOpBlockTimeout is not correctly parsed")
	}
}

func TestP2PMessagePassing(t *testing.T) {
	id := OpIdentity{clientId: "123", minerId: "345"}
	op := CreateFile{opId: id, fileName: "456", cost: 5}

	b, _ := json.Marshal(op)

	msg := Message{msgType: 3, content: b}

	fmt.Printf("Message type length %v", msg.msgType)
	c := make(chan *Message)

	go Receiver("127.0.0.1:8080", c)
	time.Sleep(2 * time.Second)
	go Sender("127.0.0.1:8081", "127.0.0.1:8080", &msg)

	receivedMsg := <- c

	var receivedOp CreateFile

	json.Unmarshal(receivedMsg.content, &receivedOp)

	if receivedOp.opId.clientId != "123" {
		t.Error("clientId is not correct")
	}

	if receivedOp.opId.minerId != "345" {
		t.Error("minerId is not correct")
	}

	if receivedOp.fileName != "456" {
		t.Error("fileName is not correct")
	}

	if receivedOp.cost != 5 {
		t.Error("cost is not correct")
	}
}

func Receiver(ipPort string, c chan *Message) {
	localTcpAddr, err := net.ResolveTCPAddr("tcp", ipPort)
	if err != nil {
		fmt.Println("Listener creation failed, please try again.")
		return
	}

	listener, err := net.ListenTCP("tcp", localTcpAddr)

	tcpConn, err := listener.AcceptTCP()
	if err != nil {
		fmt.Println("TCP connection failed.")
		return
	}

	msg, err := readMsgFromTcp(tcpConn)

	if err != nil {
		return
	}

	c <- msg
}

func Sender(ipPortLocal string, ipPortRemote string, msg *Message) {
	tcpLocalAddr, err := net.ResolveTCPAddr("tcp", ipPortLocal)
	if err != nil {
		panic("Unable to resolve local TCP address")
	}

	tcpRemoteAddr, err := net.ResolveTCPAddr("tcp", ipPortRemote)
	if err != nil {
		fmt.Println("Unable to resolve peer IpPort:", ipPortRemote)
		return
	}

	tcpConn, err := net.DialTCP("tcp", tcpLocalAddr, tcpRemoteAddr)

	if err != nil {
		fmt.Println("Failed to establish connection with peer:", ipPortRemote)
		return
	}

	sendMsgToTcp(tcpConn, msg)
}

func TestSignatureString(t *testing.T) {

	id1 := "abc"
	id2 := "def"

	c1 := 10
	c2 := 11

	sig1 := Signature{id: id1, coins: c1}
	sig2 := Signature{id: id1, coins: c1}
	sig3 := Signature{id: id2, coins: c2}

	s1 := sig1.String()
	s2 := sig2.String()
	s3 := sig3.String()

	if s1 != s2 {
		t.Error("s1 and s2 string representation should be the same")
	}

	if s2 == s3 {
		t.Error("s2 and s3 string representation should be different")
	}
}

func TestBlockHashing(t *testing.T) {

	numZeros := 5

	b := NoOpBlock{index: 1, prevHash: "abcd", sig: Signature{id: "abc", coins: 10}, nonce:0}

	b.nonce = CalcSecret(Cryptopuzzle{Hash: b.getStringWithoutNonce(), N: numZeros})

	hash := GetMd5Hash(b.String())

	zeros := strings.Repeat("0", numZeros)
	isValid := strings.HasSuffix(hash, zeros)

	if !isValid {
		t.Error("Block Hash failed, condition not satisfied")
	}
}

func TestHasEnoughCoins1(t *testing.T) {
	// First case, check balance and has enough coin
	gHash := "12344567"
	numZeros := 4
	op1 := CreateFile{opId:OpIdentity{clientId:"1", minerId:"123"}, fileName: "a", cost: 2}
	op2 := CreateFile{opId:OpIdentity{clientId:"1", minerId:"123"}, fileName: "b", cost: 3}
	op3 := CreateFile{opId:OpIdentity{clientId:"1", minerId:"345"}, fileName: "c", cost: 2}
	op4 := CreateFile{opId:OpIdentity{clientId:"1", minerId:"345"}, fileName: "d", cost: 2}
	op5 := CreateFile{opId:OpIdentity{clientId:"1", minerId:"123"}, fileName: "e", cost: 1}
	op6 := CreateFile{opId:OpIdentity{clientId:"1", minerId:"345"}, fileName: "f", cost: 2}

	opBlock1 := OpBlock{index: 1, prevHash: gHash, sig: Signature{id: "123", coins: 10}, operations: []interface{}{op1, op3}, nonce: 0}
	opBlock1.nonce = CalcSecret(Cryptopuzzle{Hash: opBlock1.getStringWithoutNonce(), N: numZeros})

	opBlock2 := OpBlock{index: 2, prevHash: opBlock1.getHash(), sig: Signature{id: "345", coins: 10}, operations: []interface{}{op2, op4}, nonce: 0}
	opBlock2.nonce = CalcSecret(Cryptopuzzle{Hash: opBlock2.getStringWithoutNonce(), N: numZeros})

	opBlock3 := OpBlock{index: 3, prevHash: opBlock2.getHash(), sig: Signature{id: "456", coins: 10}, operations: []interface{}{op5, op6}, nonce: 0}
	opBlock3.nonce = CalcSecret(Cryptopuzzle{Hash: opBlock3.getStringWithoutNonce(), N: numZeros})

	bc := BlockChain{blockChainMap: map[string] interface{}{opBlock1.getHash(): opBlock1, opBlock2.getHash(): opBlock2}, heads:[]string{opBlock2.getHash()}, initialized: true, genesisHash: gHash}

	chain := bc.getChainBeforeBlock(opBlock3)
	expectedChain := []string{opBlock1.getHash(), opBlock2.getHash()}
	if len(chain) != len(expectedChain) {
		t.Error("Different number of blocks between expected chain and actual chain")
	}
	for idx := 0; idx < len(expectedChain); idx ++ {
		if chain[idx] != expectedChain[idx] {
			t.Error("Chain before block is not generated correctly")
		}
	}


	req := opBlock3.getCoinsRequirementInBlock()

	balance := bc.getCoinBalanceInChain(chain, req, opBlock3)
	expectedBalance := map[string] int {"123": 5, "345": 6}

	if len(balance) != len(expectedBalance) {
		t.Error("Different number of items between expectedBalance and actual balance")
	}
	for minerId, expectedbal := range expectedBalance {
		actBal := balance[minerId]

		if actBal != expectedbal {
			t.Error("Expected balance:", expectedbal, "Actual balance:", actBal)
		}
	}

	hasEnoughCoins := bc.hasEnoughCoins(chain, req, opBlock3)

	if !hasEnoughCoins {
		t.Error("hasEnoughCoins returned wrong result, should have been true but got false")
	}
}
