package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
	// TODO
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

/////////// Msgs used by both auth and fortune servers:

// An error message from the server.
type ErrMessage struct {
	Error string
}

/////////// Auth server msgs:

// Message containing a nonce from auth-server.
type NonceMessage struct {
	Nonce string
	N     int64 // PoW difficulty: number of zeroes expected at end of md5(nonce+secret)
}

// Message containing an the secret value from client to auth-server.
type SecretMessage struct {
	Secret string
}

// Message with details for contacting the fortune-server.
type FortuneInfoMessage struct {
	FortuneServer string // TCP ip:port for contacting the fserver
	FortuneNonce  int64
}

/////////// Fortune server msgs:

// Message requesting a fortune from the fortune-server.
type FortuneReqMessage struct {
	FortuneNonce int64
}

// Response from the fortune-server containing the fortune.
type FortuneMessage struct {
	Fortune string
	Rank    int64 // Rank of this client solution
}

// Main workhorse method.
func main() {

	// TODO
	cmdArgs := os.Args[1:]
	localUDP := cmdArgs[0]
	localTCP := cmdArgs[1]
	aserverUDP := cmdArgs[2]

	// Use json.Marshal json.Unmarshal for encoding/decoding to servers
	addr, err := net.ResolveUDPAddr("udp", aserverUDP)
	localUDPAddr, err := net.ResolveUDPAddr("udp", localUDP)
	conn, err := net.DialUDP("udp", localUDPAddr, addr)

	if err != nil {
		fmt.Println(err)
	}

	conn.Write([]byte("hi"))
	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer)
	// Get NonceMessage from buffer

	nonceMessage := NonceMessage{Nonce: "default", N: -1}
	err = json.Unmarshal(buffer[:n], &nonceMessage)
	if err != nil {
		fmt.Println("Error", err)
	}

	// Brute force secret

	result := ""
	var secret = ""

	for !strings.HasSuffix(result, strings.Repeat("0", int(nonceMessage.N))) {
		for i := len(nonceMessage.Nonce) / 3; i < len(nonceMessage.Nonce); i++ {
			secret = RandString(i)
			result = computeNonceSecretHash(nonceMessage.Nonce, secret)
		}
	}

	// Send Secret
	secretMessage := SecretMessage{Secret: secret}
	msg, err := json.Marshal(secretMessage)

	if err != nil {
		fmt.Println("Error", err)
	}

	conn.Write(msg)

	// Receive fserver info from aserver
	buffer = make([]byte, 1024)
	n, _ = conn.Read(buffer)
	buffer = buffer[:n]

	fortune := FortuneInfoMessage{FortuneServer: "", FortuneNonce: 0}

	json.Unmarshal(buffer, &fortune)

	// step 5: send FortuneReqMessage to fserver
	requestMessage := FortuneReqMessage{FortuneNonce: fortune.FortuneNonce}

	requesetMessagePacked, err := json.Marshal(requestMessage)
	if err != nil {
		fmt.Println("Error:", err)
	}

	conn.Close()

	TCPAddr, err := net.ResolveTCPAddr("tcp", fortune.FortuneServer)
	localTCPAddr, err := net.ResolveTCPAddr("tcp", localTCP)
	TCPconn, err := net.DialTCP("tcp", localTCPAddr, TCPAddr)
	if err != nil {
		fmt.Println("Error:", err)
	}

	TCPconn.Write(requesetMessagePacked)
	buffer = make([]byte, 1024)
	n, _ = conn.Read(buffer)
	fortuneMessage := FortuneMessage{Fortune: "", Rank: -1}

	json.Unmarshal(buffer[:n], &fortuneMessage)
	fmt.Println(fortuneMessage.Fortune)
	TCPconn.Close()
}

// Returns the MD5 hash as a hex string for the (nonce + secret) value.
func computeNonceSecretHash(nonce string, secret string) string {
	h := md5.New()
	h.Write([]byte(nonce + secret))
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

var src = rand.NewSource(time.Now().UnixNano())

func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
