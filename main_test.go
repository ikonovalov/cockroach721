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
	"crypto/ecdsa"
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

func createContext() context.Context {
	return context.Background()
}

func createKeysAndTransactOpts() (*ecdsa.PrivateKey, *bind.TransactOpts) {
	key, _ := crypto.GenerateKey()
	opts := bind.NewKeyedTransactor(key)
	return key, opts
}

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
	fmt.Println("== TestDeployBreedingNFToken =======================================================================")
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
	fmt.Println(" TestSpawnOne ======================================================================================")
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
	fmt.Println(" TestSpawnTenCockroachAndHandleEvent ===============================================================")
	var (
		ctx                = createContext()
		_, auth            = createKeysAndTransactOpts()
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

func TestSetSpeedFeeAndSpawn(t *testing.T) {
	fmt.Println(" TestSetSpeedFeeAndSpawn ==========================================================================")
	var (
		ctx                = createContext()
		_, auth            = createKeysAndTransactOpts()
		backend            = createSimulatedBackend(auth)
		suggestGasPrice, _ = backend.SuggestGasPrice(ctx)
		breedingToken      = deployWith(ctx, backend, auth)
	)

	tx, _ := breedingToken.SetSpeedUnitFee(auth, new(big.Int).SetUint64(13*finney))
	backend.Commit()
	waitTx(backend, ctx, tx)

	var (
		newSpeedFee uint64 = 13
		speed       uint64 = 5
	)


	if fee, _ := breedingToken.SpeedUnitFee(nil); fee.Uint64() != (newSpeedFee * finney) {
		t.Error("Wrong fee. Expected 13 finney.")
	} else {
		fmt.Printf("New speed fee is %d\n", fee)
	}

	// Best fit options
	spawnOps := &bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		Value:    new(big.Int).SetUint64(newSpeedFee * speed * finney),
		GasPrice: suggestGasPrice,
		GasLimit: uint64(200000),
	}

	tx, _ = breedingToken.Spawn(spawnOps, "Mary", 5)
	backend.Commit()

	if txReceipt := waitTx(backend, ctx, tx); txReceipt.CumulativeGasUsed < 50000 {
		t.Errorf("Something goes wrong. OGG? We use too low gas %d\n", txReceipt.CumulativeGasUsed)
	} else {
		fmt.Printf("Spawn gas used %d\n", txReceipt.CumulativeGasUsed)
	}

	// Too low options
	spawnOpsTooLow := &bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		Value:    new(big.Int).SetUint64(newSpeedFee*speed*finney - 1),
		GasPrice: suggestGasPrice,
		GasLimit: uint64(200000),
	}
	tx, _ = breedingToken.Spawn(spawnOpsTooLow, "Mary never", 5)
	backend.Commit()

	if txReceipt := waitTx(backend, ctx, tx); txReceipt.CumulativeGasUsed > 30000 {
		fmt.Errorf("It should raise OOG. But used gas is %d\n", txReceipt.CumulativeGasUsed)
	}

}

func TestSpawnCockroachAndView(t *testing.T) {
	fmt.Println(" TestSpawnCockroachAndView ===============================================================")
	var (
		ctx                = createContext()
		_, auth            = createKeysAndTransactOpts()
		backend            = createSimulatedBackend(auth)
		suggestGasPrice, _ = backend.SuggestGasPrice(ctx)
		breedingToken      = deployWith(ctx, backend, auth)
	)
	addr := auth.From
	fmt.Printf("Coinbase %s\n", addr.Hex())

	spawnOps := &bind.TransactOpts{
		From:     addr,
		Signer:   auth.Signer,
		Value:    new(big.Int).SetUint64(5 * finney),
		GasPrice: suggestGasPrice,
		GasLimit: uint64(200000),
	}

	tx, err := breedingToken.Spawn(spawnOps, "Mary", 5)
	failIf(err, t)
	backend.Commit()
	waitTx(backend, ctx, tx)


	if balanceOf, _ := breedingToken.BalanceOf(nil, addr); balanceOf.Uint64() != 1 {
		t.FailNow()
	}

	tokens, _ := breedingToken.GetOwnerTokens(nil, addr)
	for _, tokenId := range tokens {
		currentCockroach, _ := breedingToken.Cockroaches(nil, tokenId)
		fmt.Println(currentCockroach)
	}
}
