package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type cryptopuzzle struct {
	Hash string // block hash without nonce
	N    int    // PoW difficulty: number of zeroes expected at end of md5
}

func calcSecret(problem cryptopuzzle) (nonce string) {
	result := ""

	for !validNonce(problem.N, result) {
		nonce = randString()
		result = computeNonceSecretHash(problem.Hash, nonce)
	}
	return
}

func computeNonceSecretHash(hash string, nonce string) string {
	h := md5.New()
	h.Write([]byte(hash + nonce))
	str := hex.EncodeToString(h.Sum(nil))
	return str
}

func validNonce(N int, Hash string) bool {
	zeros := strings.Repeat("0", N)
	isValid := strings.HasSuffix(Hash, zeros)
	fmt.Println("valid: " + Hash)
	return isValid
}

func randString() string {
	// tradeoff
	// the larger the possible max length of string the longer it takes to generate each string
	// pro: more possible solutions to try
	// con: much slower in each try
	n := rand.Intn(1000)
	output := make([]byte, n)
	// We will take n bytes, one byte for each character of output.
	randomness := make([]byte, n)
	// read all random
	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(letterBytes)
	// fill output
	for pos := range output {
		// get random item
		random := uint8(randomness[pos])
		// random % 64
		randomPos := random % uint8(l)
		// put into output
		output[pos] = letterBytes[randomPos]
	}
	return string(output)
}
