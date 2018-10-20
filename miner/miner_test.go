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

	c1 := uint32(10)
	c2 := uint32(11)

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
