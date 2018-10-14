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

