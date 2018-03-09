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
	"github.com/ethereum/go-ethereum/common"
)

// 1 Ether = 1000000000000000000

var (
	wei     = uint64(math.Pow(10.0, 0.0))
	ada     = uint64(math.Pow(10.0, 3.0))
	babbage = uint64(math.Pow(10.0, 6.0))
	shannon = uint64(math.Pow(10.0, 9.0))
	szabo   = uint64(math.Pow(10.0, 12.0))
	finney  = uint64(math.Pow(10.0, 15.0))
	ether   = uint64(math.Pow(10.0, 18.0))
)

func createSimulatedBackend(auth *bind.TransactOpts) *backends.SimulatedBackend {
	alloc := make(core.GenesisAlloc)
	alloc[auth.From] = core.GenesisAccount{
		Balance: new(big.Int).SetUint64(1 * ether),
	}
	sim := backends.NewSimulatedBackend(alloc)
	return sim
}

func deployWith(ctx context.Context, backend *backends.SimulatedBackend, txOps *bind.TransactOpts) (breedingToken *cockroach.CockroachToken) {
	_, tx, breedingToken, err := cockroach.DeployCockroachToken(txOps, backend)
	exitIf(err)
	backend.Commit()
	waitDeployed(backend, ctx, tx)
	receipt, _ := backend.TransactionReceipt(ctx, tx.Hash())
	fmt.Printf("Deploy gas used %d.\n", receipt.CumulativeGasUsed)
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
	suggestGasPrice, _ := backend.SuggestGasPrice(ctx)

	// deploy breeding token
	breedingToken := deployWith(ctx, backend, auth)

	balanceAtInitial, err := backend.BalanceAt(ctx, auth.From, nil)
	fmt.Printf("Account %s has balance %s wei.\n", auth.From.Hex(), balanceAtInitial.String())
	failIf(err, t)

	// Initial total supply
	supply, err := breedingToken.TotalSupply(nil)
	failIf(err, t)
	fmt.Printf("Initial population %d.\n", supply.Int64())
	if supply.Int64() != 0 {
		t.Error("Wrong initial population. Not zero.")
	}

	spawnGasLimit := uint64(200000)

	spawnOps := &bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		Value:    new(big.Int).SetUint64(5*finney + 17),
		GasPrice: suggestGasPrice,
		GasLimit: spawnGasLimit,
	}
	tx, err := breedingToken.Spawn(spawnOps, "Jack", 5)
	failIf(err, t)
	backend.Commit()

	txReceipt := waitTx(backend, ctx, tx)
	if txReceipt.CumulativeGasUsed == spawnGasLimit {
		t.Error("OOG!")
	} else {
		fmt.Printf("SPAWN: TxReceipt %s GasUsed %d. GasLimit %d.\n", txReceipt.TxHash.Hex(), txReceipt.CumulativeGasUsed, spawnOps.GasLimit)
	}

	supply, err = breedingToken.TotalSupply(nil)
	failIf(err, t)
	if supply.Int64() != 1 {
		t.Errorf("Wrong population count after one spawn op. Expected 1, but got %s.", supply)
	}

	// check final balance
	balanceAtEnd, _ := backend.BalanceAt(ctx, auth.From, nil)
	fmt.Printf("Account %s has balance %s wei.\n", auth.From.Hex(), balanceAtEnd.String())
	balanceDiff := new(big.Int).Sub(balanceAtInitial, balanceAtEnd).Uint64()

	usedGasFee := suggestGasPrice.Uint64() * txReceipt.CumulativeGasUsed
	fmt.Printf("Used gas fee %d wei.\n", usedGasFee)
	if balanceDiff != (5*finney + usedGasFee) {
		t.Error("Bad ending balance")
	}
}

func TestSpawnTenCockroachAndHandleEvent(t *testing.T) {
	fmt.Println("===================================================================================================")
	var (
		ctx                = context.Background()
		key, _             = crypto.GenerateKey()
		auth               = bind.NewKeyedTransactor(key)
		backend            = createSimulatedBackend(auth)
		suggestGasPrice, _ = backend.SuggestGasPrice(ctx)
		breedingToken      = deployWith(ctx, backend, auth)
	)
	fmt.Printf("Coinbase %s\n", auth.From.Hex())

	spawnOps := &bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		Value:    new(big.Int).SetUint64(5 * finney),
		GasPrice: suggestGasPrice,
		GasLimit: uint64(200000),
	}

	for i := 0; i < 10; i++ {
		tx, err := breedingToken.Spawn(spawnOps, "Mary", 5)
		failIf(err, t)
		backend.Commit()
		waitTx(backend, ctx, tx)
	}

	iterator, err := breedingToken.FilterSpawn(new(bind.FilterOpts), []common.Address{auth.From})
	failIf(err, t)
	eventCount := 0
	for iterator.Next() {
		var (
			spawnEvent = iterator.Event
			tokenId    = spawnEvent.TokenId.Uint64()
			owner      = spawnEvent.To.Hex()
		)
		fmt.Printf("SPAWN EVENT. TokenID %d, OWNER %s\n", tokenId, owner)
		eventCount++
	}

	if eventCount != 10 {
		t.FailNow()
	}
}
