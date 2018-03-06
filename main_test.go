package main

import (
	"testing"
	"context"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ikonovalov/go-cockroach/contracts/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"math/big"
	"fmt"
	"math"
	"github.com/ethereum/go-ethereum/core"
)

// 1 Ether = 1000000000000000000

var (
	wei     = int64(math.Pow(10.0, 0.0))
	ada     = int64(math.Pow(10.0, 3.0))
	babbage = int64(math.Pow(10.0, 6.0))
	shannon = int64(math.Pow(10.0, 9.0))
	szabo   = int64(math.Pow(10.0, 12.0))
	finney  = int64(math.Pow(10.0, 15.0))
	ether   = int64(math.Pow(10.0, 18.0))
)

func createSimulatedBackend(auth *bind.TransactOpts) *backends.SimulatedBackend {
	alloc := make(core.GenesisAlloc)
	alloc[auth.From] = core.GenesisAccount{
		Balance: big.NewInt(1 * ether),
	}
	sim := backends.NewSimulatedBackend(alloc)
	return sim
}

func deployWith(ctx context.Context, backend *backends.SimulatedBackend, txOps *bind.TransactOpts) (breedingToken *cockroach.CockroachToken) {
	_, tx, breedingToken, err := cockroach.DeployCockroachToken(txOps, backend)
	exitIf(err)
	backend.Commit()
	waitDeployed(backend, ctx, tx)
	return
}

func failIf(err error, t *testing.T) {
	if err != nil {
		t.Error(err)
	}
}

func TestDeployBreedingNFToken(t *testing.T) {
	fmt.Println("===================================================================================================")
	ctx := context.Background()
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	backend := createSimulatedBackend(auth)

	_, tx, _, err := cockroach.DeployCockroachToken(auth, backend)
	exitIf(err)

	backend.Commit()
	waitDeployed(backend, ctx, tx)
}

func TestSpawnOne(t *testing.T) {
	fmt.Println("===================================================================================================")
	ctx := context.Background()
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	backend := createSimulatedBackend(auth)

	breedingToken := deployWith(ctx, backend, auth)

	balanceAt, err := backend.BalanceAt(ctx, auth.From, nil)
	fmt.Printf("Account %s has balance %s wei.\n", auth.From.Hex(), balanceAt.String())
	failIf(err, t)

	// Initial total supply
	supply, err := breedingToken.TotalSupply(nil)
	failIf(err, t)
	fmt.Printf("Initial population %d.\n", supply.Int64())
	if supply.Int64() != 0 {
		t.Error("Wrong initial population. Not zero.")
	}

	//speedPriceRequired, _ := breedingToken.CalcSpeedPrice(nil, 5)
	//fmt.Printf("Required speed fee: %d wei.\n", speedPriceRequired.Int64())

	spawnGasLimit := uint64(200000)

	spawnOps := &bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		Value:    big.NewInt(5 * 1000000000000000),
		GasLimit: spawnGasLimit,
	}
	tx, err := breedingToken.Spawn(spawnOps, "Jack", 5)
	failIf(err, t)
	backend.Commit()

	txReceipt := waitTx(backend, ctx, tx)
	if txReceipt.GasUsed == spawnGasLimit {
		t.Error("Throw or OOG!")
	} else {
		fmt.Printf("SPAWN: TxReceipt %s GasUsed %d. GasLimit %d.\n", txReceipt.TxHash.Hex(), txReceipt.GasUsed, spawnOps.GasLimit)
	}

	supply, err = breedingToken.TotalSupply(nil)
	failIf(err, t)
	if supply.Int64() != 1 {
		t.Errorf("Wrong population count after one spawn op. Expected 1, but got %s.", supply)
	}
}
