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
)

// 1 Ether = 1000000000000000000

func deployWith(ctx context.Context, txOps *bind.TransactOpts) (breedingToken *cockroach.CockroachToken, backend *backends.SimulatedBackend) {
	backend = createSimulatedBackend(txOps)
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
	ctx := context.Background()
	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)

	breedingToken, backend := deployWith(ctx, auth)

	balanceAt, err := backend.BalanceAt(ctx, auth.From, nil)
	fmt.Printf("%s balance %s\n", auth.From.Hex(), balanceAt.String())
	failIf(err, t)

	supply, err := breedingToken.TotalSupply(nil)
	failIf(err, t)

	if supply.Int64() != 0 {
		t.Error("Wrong initial population. Not zero.")
	}

	spawnGasLimit := uint64(200000)

	spawnOps := &bind.TransactOpts{
		From:     auth.From,
		Signer:   auth.Signer,
		Value:    big.NewInt(0),
		GasLimit: spawnGasLimit,
	}
	tx, err := breedingToken.Spawn(spawnOps, "Jack", 5)
	failIf(err, t)
	backend.Commit()

	txReceipt := waitTx(backend, ctx, tx)
	if txReceipt.GasUsed == spawnGasLimit {
		t.Error("Throw or OOG!")
	} else {
		fmt.Printf("Spawn op use %d gas\n", txReceipt.GasUsed)
	}

	supply, err = breedingToken.TotalSupply(nil)
	failIf(err, t)
	if supply.Int64() != 1 {
		t.Errorf("Wrong population count after one spawn op. Expected 1, but got %s.", supply)
	}
}
