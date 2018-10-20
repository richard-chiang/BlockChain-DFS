package tests

import (
	miner2 "cpsc416-p1/miner"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestMinerInitialization(t *testing.T) {
	jsonPath := "miner1.json"
	miner, err := miner2.InitializeMiner(jsonPath)

	if err != nil {
		fmt.Println(err)
		t.Errorf("error incured when initializing the miner.")
	}

	expectedPerrs := []string{"123.123.123.123:100", "234.234.234.234:200", "345,345,345.345:300"}

	for idx := 0; idx < len(expectedPerrs); idx++ {
		if expectedPerrs[idx] != miner.Config.PeerMinersAddrs[idx] {
			t.Errorf("Field PeerMinersAddrs is not corrected parsed.")
		}
	}

	if miner.Config.IncomingMinersAddr != "127.0.0.1:8081" {
		t.Errorf("Field IncomingMinersAddr is not corrected parsed.")
	}

	if miner.Config.OutgoingMinersIP != "127.0.0.1:8082" {
		t.Errorf("Field OutgoingMinersIP is not corrected parsed.")
	}

	if miner.Config.IncomingClientsAddr != "127.0.0.1:8080" {
		t.Errorf("Field IncomingClientsAddr is not corrected parsed.")
	}

	if miner.Config.MinerID != "123456" {
		t.Errorf("Field MinerID is not correctly parsed.")
	}

	if miner.Config.MinedCoinsPerOpBlock != 2 {
		t.Errorf("Field MinedCoinsPerOpBlock is not correctly parsed")
	}

	if miner.Config.MinedCoinsPerNoOpBlock != 3 {
		t.Errorf("Field MinedCoinsPerNoOpBlock is not correctly parsed")
	}

	if miner.Config.NumCoinsPerFileCreate != 1 {
		t.Errorf("Field NumCoinsPerFileCreate is not correctly parsed")
	}

	if miner.Config.GenesisBlockHash != "qwerwerqwfsdfsadfsdf" {
		t.Errorf("Field GenesisBlockHash is not correctly parsed")
	}

	if miner.Config.PowPerOpBlock != 6 {
		t.Error("Field PowPerOpBlock is not correctly parsed")
	}

	if miner.Config.PowPerNoOpBlock != 5 {
		t.Error("Field PowPerNoOpBlock is not correctly parsed")
	}

	if miner.Config.ConfirmsPerFileCreate != 5 {
		t.Error("Field ConfirmsPerFileCreate is not correctly parsed")
	}

	if miner.Config.ConfirmsPerFileAppend != 6 {
		t.Error("Field ConfirmsPerFileAppend is not correctly parsed")
	}

	if miner.Config.GenOpBlockTimeout != 5 {
		t.Error("Field GenOpBlockTimeout is not correctly parsed")
	}
}

func TestP2PMessagePassing(t *testing.T) {
	id := miner2.OpIdentity{ClientId: "123", MinerId: "345"}
	op := miner2.CreateFile{OpId: id, FileName: "456", Cost: 5}

	b, _ := json.Marshal(op)

	msg := miner2.Message{Type: 3, Content: b}

	fmt.Printf("Message type length %v", msg.Type)
	c := make(chan *miner2.Message)

	go Receiver("127.0.0.1:8080", c)
	time.Sleep(2 * time.Second)
	go Sender("127.0.0.1:8081", "127.0.0.1:8080", &msg)

	receivedMsg := <- c

	var receivedOp miner2.CreateFile

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

func Receiver(ipPort string, c chan *miner2.Message) {
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

	msg, err := miner2.ReadMsgFromTcp(tcpConn)

	if err != nil {
		return
	}

	c <- msg
}

func Sender(ipPortLocal string, ipPortRemote string, msg *miner2.Message) {
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

	miner2.SendMsgToTcp(tcpConn, msg)
}

func TestSignatureString(t *testing.T) {

	id1 := "abc"
	id2 := "def"

	c1 := uint32(10)
	c2 := uint32(11)

	sig1 := miner2.Signature{Id: id1, Coins: c1}
	sig2 := miner2.Signature{Id: id1, Coins: c1}
	sig3 := miner2.Signature{Id: id2, Coins: c2}

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

	b := miner2.NoOpBlock{Index: 1, PrevHash: "abcd", Sig: miner2.Signature{Id: "abc", Coins: 10}, Nonce:0}

	b.Nonce = miner2.CalcSecret(miner2.Cryptopuzzle{Hash: b.GetStringWithoutNonce(), N: numZeros})

	hash := miner2.GetMd5Hash(b.String())

	zeros := strings.Repeat("0", numZeros)
	isValid := strings.HasSuffix(hash, zeros)

	if !isValid {
		t.Error("Block hash failed, condition not satisfied")
	}
}
