package tests

import (
	miner2 "cpsc416-p1/miner"
	"fmt"
	"testing"
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
		if expectedPerrs[idx] != miner.Config.Peers[idx] {
			t.Errorf("Field Peers is not corrected parsed.")
		}
	}

	if miner.Config.LocalTcpIpPort != "127.0.0.1:8081" {
		t.Errorf("Field LocalTcpIpPort is not corrected parsed.")
	}

	if miner.Config.ClientTcpIpPort != "127.0.0.1:8080" {
		t.Errorf("Field ClientTcpIpPort is not corrected parsed.")
	}

	if miner.Config.Id != "123456" {
		t.Errorf("Field miner Id is not correctly parsed.")
	}

	if miner.Config.NumCoinsOpBlock != 2 {
		t.Errorf("Field NumCoinsOpBlock is not correctly parsed")
	}

	if miner.Config.NumCoinsNopBlock != 3 {
		t.Errorf("Field NumCoinsNopBlock is not correctly parsed")
	}

	if miner.Config.NumCoinsPerFile != 1 {
		t.Errorf("Field NumCoinsPerFile is not correctly parsed")
	}

	if miner.Config.GenesisHash != "qwerwerqwfsdfsadfsdf" {
		t.Errorf("Field GenesisHash is not correctly parsed")
	}

	if miner.Config.OpDifficulty != 6 {
		t.Error("Field OpDifficulty is not correctly parsed")
	}

	if miner.Config.NopDifficulty != 5 {
		t.Error("Field NopDifficulty is not correctly parsed")
	}

	if miner.Config.NumConfFc != 5 {
		t.Error("Field NumConfFc is not correctly parsed")
	}

	if miner.Config.NumConfApp != 6 {
		t.Error("Field NumConfApp is not correctly parsed")
	}
}