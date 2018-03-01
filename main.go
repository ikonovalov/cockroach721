package main

import (
	"github.com/ikonovalov/go-cockroach/contracts/bind"
	"math/big"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"log"
	"fmt"
	"context"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	ctx := context.Background()

	key, _ := crypto.GenerateKey()
	auth := bind.NewKeyedTransactor(key)
	fmt.Printf("Coinbase: %s\n", auth.From.Hex())

	backend := createSimulatedBackend(auth)

	_, tx, token, err := cockroach.DeployCockroachToken(auth, backend)
	exitIf(err)

	backend.Commit()
	waitDeployed(backend, ctx, tx)

	name, _ := token.Name(nil)
	symbol, _ := token.Symbol(nil)
	totalSupply, _ := token.TotalSupply(nil)
	fmt.Printf("Name: %s\n", name)
	fmt.Printf("Symbol: %s\n", symbol)
	fmt.Printf("Total: %s\n", totalSupply)

}

func createSimulatedBackend(auth *bind.TransactOpts) *backends.SimulatedBackend {
	alloc := make(core.GenesisAlloc)
	alloc[auth.From] = core.GenesisAccount{Balance: big.NewInt(13370000000)}
	sim := backends.NewSimulatedBackend(alloc)
	return sim
}

func exitIf(e error) {
	if e != nil {
		log.Fatal(e)
	}
}