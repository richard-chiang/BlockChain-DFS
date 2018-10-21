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
	id := OpIdentity{ClientId: "123", MinerId: "345"}
	op := CreateFile{OpId: id, FileName: "456", Cost: 5}

	b, _ := json.Marshal(op)

	nowTime := time.Now()
	msg := Message{MsgType: 3, T: nowTime, Content: b}

	fmt.Printf("Message type length %v\n", msg.MsgType)
	c := make(chan *Message)

	go Receiver("127.0.0.1:8080", c)
	time.Sleep(2 * time.Second)
	go Sender("127.0.0.1:8081", "127.0.0.1:8080", &msg)

	receivedMsg := <- c

	var receivedOp CreateFile

	json.Unmarshal(receivedMsg.Content, &receivedOp)

	if receivedOp.OpId.ClientId != "123" {
		t.Error("ClientId is not correct")
	}

	if receivedOp.OpId.MinerId != "345" {
		t.Error("MinerId is not correct")
	}

	if receivedOp.FileName != "456" {
		t.Error("FileName is not correct")
	}

	if receivedOp.Cost != 5 {
		t.Error("Cost is not correct")
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

	sig1 := Signature{Id: id1, Coins: c1}
	sig2 := Signature{Id: id1, Coins: c1}
	sig3 := Signature{Id: id2, Coins: c2}

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

	b := NoOpBlock{Index: 1, PrevHash: "abcd", Sig: Signature{Id: "abc", Coins: 10}, Nonce:0}

	b.Nonce = CalcSecret(Cryptopuzzle{Hash: b.getStringWithoutNonce(), N: numZeros})

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
	op1 := CreateFile{OpId: OpIdentity{ClientId: "1", MinerId:"123"}, FileName: "a", Cost: 2}
	op2 := CreateFile{OpId: OpIdentity{ClientId: "1", MinerId:"123"}, FileName: "b", Cost: 3}
	op3 := CreateFile{OpId: OpIdentity{ClientId: "1", MinerId:"345"}, FileName: "c", Cost: 2}
	op4 := CreateFile{OpId: OpIdentity{ClientId: "1", MinerId:"345"}, FileName: "d", Cost: 2}
	op5 := CreateFile{OpId: OpIdentity{ClientId: "1", MinerId:"123"}, FileName: "e", Cost: 1}
	op6 := CreateFile{OpId: OpIdentity{ClientId: "1", MinerId:"345"}, FileName: "f", Cost: 2}

	opBlock1 := OpBlock{Index: 1, PrevHash: gHash, Sig: Signature{Id: "123", Coins: 10}, Operations: []interface{}{op1, op3}, Nonce: 0}
	opBlock1.Nonce = CalcSecret(Cryptopuzzle{Hash: opBlock1.getStringWithoutNonce(), N: numZeros})

	opBlock2 := OpBlock{Index: 2, PrevHash: opBlock1.getHash(), Sig: Signature{Id: "345", Coins: 10}, Operations: []interface{}{op2, op4}, Nonce: 0}
	opBlock2.Nonce = CalcSecret(Cryptopuzzle{Hash: opBlock2.getStringWithoutNonce(), N: numZeros})

	opBlock3 := OpBlock{Index: 3, PrevHash: opBlock2.getHash(), Sig: Signature{Id: "456", Coins: 10}, Operations: []interface{}{op5, op6}, Nonce: 0}
	opBlock3.Nonce = CalcSecret(Cryptopuzzle{Hash: opBlock3.getStringWithoutNonce(), N: numZeros})

	bc := BlockChain{BlockChainMap: map[string] interface{}{opBlock1.getHash(): opBlock1, opBlock2.getHash(): opBlock2}, Heads:[]string{opBlock2.getHash()}, Initialized: true, GenesisHash: gHash}

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
